package log

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	api "github.com/jhkim988/proglog/api/v1"
	"google.golang.org/protobuf/proto"
)

type DistributedLog struct {
	config  Config
	log     *Log
	raftLog *logStore
	raft    *raft.Raft
}

func NewDistributedLog(dataDir string, config Config) (*DistributedLog, error) {
	l := &DistributedLog{
		config: config,
	}

	if err := l.setupLog(dataDir); err != nil {
		return nil, err
	}

	if err := l.setupRaft(dataDir); err != nil {
		return nil, err
	}
	return l, nil
}

// 서버에 로그를 생성한다. 서버는 이 로그에 사용자의 레코드를 저장한다.
func (l *DistributedLog) setupLog(dataDir string) error {
	logDir := filepath.Join(dataDir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	var err error
	l.log, err = NewLog(logDir, l.config)
	return err
}

/*
래프트 인스턴트 구성요소
1. 래프트에 준 명령을 적용하는 유한상태머신
2. 래프트가 명령을 저장하는 로그 저장소
3. 클러스터 내 서버들과 그 주소 등의 메타데이터(설정)을 저장하는 stable store
4. snapshot 저장소
5. 다른 서버에 연결할 때 사용하는 transport
*/
func (l *DistributedLog) setupRaft(dataDir string) error {
	fsm := &fsm{log: l.log}

	/* 로그 저장소 설정 */
	logDir := filepath.Join(dataDir, "raft", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	logConfig := l.config
	logConfig.Segment.InitialOffset = 1

	var err error
	l.raftLog, err = newLogStore(logDir, logConfig)
	if err != nil {
		return err
	}

	/* stable store 설정, Bolt: 내장형 key-value db */
	stableStore, err := raftboltdb.NewBoltStore(
		filepath.Join(dataDir, "raft", "stable"),
	)
	if err != nil {
		return err
	}

	/* snapshot store 설정 */
	/* 새로운 인스턴스가 래프트 리더로부터 모든 데이터를 스트리밍 받는 것보다 스냅숏에서 복원하는 편이 효율적이다. */
	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(dataDir, "raft"),
		retain,
		os.Stderr,
	)
	if err != nil {
		return err
	}

	/* transport 설정 */
	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
		l.config.Raft.StreamLayer,
		maxPool,
		timeout,
		os.Stderr,
	)

	config := raft.DefaultConfig()
	config.LocalID = l.config.Raft.LocalID
	if l.config.Raft.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = l.config.Raft.HeartbeatTimeout
	}
	if l.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = l.config.Raft.ElectionTimeout
	}
	if l.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = l.config.Raft.LeaderLeaseTimeout
	}
	if l.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = l.config.Raft.CommitTimeout
	}

	/* raft 인스턴스 생성 */
	l.raft, err = raft.NewRaft(
		config,
		fsm,
		l.raftLog,
		stableStore,
		snapshotStore,
		transport,
	)
	if err != nil {
		return err
	}

	hasState, err := raft.HasExistingState(
		l.raftLog,
		stableStore,
		snapshotStore,
	)
	if err != nil {
		return err
	}

	/* 클러스터 부트스트랩 실행 */
	/* 부트스트랩은 서버 자신을 유일한 투표자로 설정하고 리더가 될 때까지 기다린 다음, 리더가 더 많은 서버를 클러스터에 추가하도록 한다.*/
	if l.config.Raft.Bootstrap && !hasState {
		config := raft.Configuration{
			Servers: []raft.Server{{
				ID:      config.LocalID,
				Address: transport.LocalAddr(),
			}},
		}
		err = l.raft.BootstrapCluster(config).Error()
	}
	return err
}

/* DistributedLog 구조체는 Log 구조체와 같은 API 를 가지도록 하여, 서로 호환되도록 한다. */
func (l *DistributedLog) Append(record *api.Record) (uint64, error) {
	/* 서버의 로그에 직접 추가하지 않고, FSM 이 레코드를 로그에 추가하도록 한다. */
	res, err := l.apply(
		AppendRequestType,
		&api.ProduceRequest{Record: record},
	)
	if err != nil {
		return 0, err
	}
	return res.(*api.ProduceResponse).Offset, nil
}

/* raft API 를 감싸고, API 응답을 리턴한다. */
func (l *DistributedLog) apply(reqType RequestType, req proto.Message) (interface{}, error) {
	// 요청을 직렬화하여 byte 로 만든다.
	var buf bytes.Buffer

	_, err := buf.Write([]byte{byte(reqType)})
	if err != nil {
		return nil, err
	}

	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}

	// 로그복제를 수행하고 리더의 로그에 레코드를 추가한다.
	timeout := 10 * time.Second
	future := l.raft.Apply(buf.Bytes(), timeout)
	// Error(): raft 복제가 잘못되었을 때 에러를 리턴한다. (수행시간이 너무 오래 걸리거나, 정지해야할 때)
	if future.Error() != nil {
		return nil, future.Error()
	}

	// Response(): FSM 의 Apply() 메서드가 리턴하는 것을 받아 리턴한다.
	res := future.Response()
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res, nil
}

func (l *DistributedLog) Read(offset uint64) (*api.Record, error) {
	return l.log.Read(offset)
}

// fsm 이 raft.FSM 인터페이스를 만족하는지 확인.
var _ raft.FSM = (*fsm)(nil)

/*
	FSM
	1. Apply: raft 는 로그를 커밋하면 이 메서드를 호출한다.
	2. Snapshot: 상태 snapshot 을 찍고자 주기적으로 호출한다. 최적화된 로그를 만들 수 있다. (a -> b -> c 로 변경되면 마지막 변경사항만 저장)
	3. Restore: snapshot 으로 FSM 을 복원할 때 호출한다. 새로운 인스턴스를 생성한 경우 이 메서드를 이용해 복원한다.
*/

type fsm struct {
	log *Log
}

type RequestType uint8

// 여러 명령을 지원하도록 구현하려면, RequestType 상수를 추가한다.
const (
	AppendRequestType RequestType = 0
)

func (l *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	reqType := RequestType(buf[0])

	// RequestType 에 따라 처리한다.
	switch reqType {
	case AppendRequestType:
		return l.applyAppend(buf[1:])
	}
	return nil
}

func (l *fsm) applyAppend(b []byte) interface{} {
	// 요청을 역직렬화한다.
	var req api.ProduceRequest
	err := proto.Unmarshal(b, &req)
	if err != nil {
		return err
	}

	// 로그에 추가한다.
	offset, err := l.log.Append(req.Record)
	if err != nil {
		return err
	}
	return &api.ProduceResponse{Offset: offset}
}

/* FSM의 상태에 대한 특정 시점의 snapshot 을 리턴한다. */
/* raft 는 snapshot 을 찍을 시간을 체크하는 SnapshotInterval 설정과, 마지막 snapshot 이후 추가한 로그 개수인 SnapshotThreshold 설정에 따라 Snpashot 메서드를 호출한다. */
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	// io.Reader 를 리턴하여 모든 로그 데이터를 읽을 수 있게 한다.
	r := f.log.Reader()
	return &snapshot{reader: r}, nil
}

/* snapshot 이 raft.FSMSnapShot 인터페이스를 만족하는지 확인하는 코드 */
var _ raft.FSMSnapshot = (*snapshot)(nil)

type snapshot struct {
	reader io.Reader
}

/* FSMSnapshot 에 Persist 를 호출하여 상태를 sink 에 쓰도록 한다. */
/* sink 는 snapshot 의 저장소, 인메모리, 파일, S3 bucket 등을 설정할 수 있다. */
func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

/* Snapshot 을 찍고나면 Release 를 호출한다. */
func (s *snapshot) Release() {}

/* 기존의 상태를 없애고, 리더의 복제 상태와 똑같아지도록 한다. */
func (f *fsm) Restore(r io.ReadCloser) error {
	b := make([]byte, lenWidth)
	var buf bytes.Buffer

	for i := 0; ; i++ {
		_, err := io.ReadFull(r, b)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		size := int64(enc.Uint64(b))
		if _, err = io.CopyN(&buf, r, size); err != nil {
			return err
		}

		record := &api.Record{}
		if err = proto.Unmarshal(buf.Bytes(), record); err != nil {
			return err
		}

		if i == 0 {
			f.log.Config.Segment.InitialOffset = record.Offset
			if err := f.log.Reset(); err != nil {
				return err
			}
		}

		if _, err = f.log.Append(record); err != nil {
			return err
		}

		buf.Reset()
	}
	return nil
}

/* raft 는 managed 로그 저장소에서 *raft.Log 를 읽어, FSM 의 Apply() 에 넣는다. */
/* logStore 가 raft.LogStore 인터페이스를 만족하는지 확인 */
var _ raft.LogStore = (*logStore)(nil)

type logStore struct {
	*Log
}

func newLogStore(dir string, c Config) (*logStore, error) {
	log, err := NewLog(dir, c)
	if err != nil {
		return nil, err
	}
	return &logStore{log}, nil
}

func (l *logStore) FirstIndex() (uint64, error) {
	return l.LowestOffset()
}

func (l *logStore) LastIndex() (uint64, error) {
	off, err := l.HighestOffset()
	return off, err
}

func (l *logStore) GetLog(index uint64, out *raft.Log) error {
	in, err := l.Read(index)
	if err != nil {
		return err
	}
	out.Data = in.Value
	out.Index = in.Offset
	out.Type = raft.LogType(in.Type)
	out.Term = in.Term
	return nil
}

func (l *logStore) StoreLog(record *raft.Log) error {
	return l.StoreLogs([]*raft.Log{record})
}

func (l *logStore) StoreLogs(records []*raft.Log) error {
	for _, record := range records {
		if _, err := l.Append(&api.Record{
			Value: record.Data,
			Term:  record.Term,
			Type:  uint32(record.Type),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (l *logStore) DeleteRange(min, max uint64) error {
	return l.Truncate(max)
}

// raft 는 스트림 계층을 전송에 사용하여 추상화를 제공하고, 래프트 서버들을 연결한다.
// StreamLayer 구조체를 정의하고, raft.StreamLayer 인터페이스를 만족하는지 확인한다.
var _ raft.StreamLayer = (*StreamLayer)(nil)

type StreamLayer struct {
	ln              net.Listener
	serverTLSConfig *tls.Config
	peerTLSConfig   *tls.Config
}

func NewStreamLayer(ln net.Listener, serverTLSConfig, peerTLSConfig *tls.Config) *StreamLayer {
	return &StreamLayer{
		ln:              ln,
		serverTLSConfig: serverTLSConfig, // 연결요청을 수락
		peerTLSConfig:   peerTLSConfig,   // 외부로의 연결을 생성
	}
}

const RaftRPC = 1

// Raft 클러스트의 다른 서버와 연결한다.
func (s *StreamLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	var conn, err = dialer.Dial("tcp", string(addr))
	if err != nil {
		return nil, err
	}

	// mux 에 raft rpc 임을 알린다.
	// 연결 자료형을 명시해서 raft 가 Log gRPC 요청을 보내는 포토를 함께 사용
	_, err = conn.Write([]byte{byte(RaftRPC)})
	if err != nil {
		return nil, err
	}

	// 스트림계층에서 peerTLS 설정 -> TLS 클라이언트 연결
	if s.peerTLSConfig != nil {
		conn = tls.Client(conn, s.peerTLSConfig)
	}
	return conn, err
}

// 요청을 수락한다.
func (s *StreamLayer) Accept() (net.Conn, error) {
	conn, err := s.ln.Accept()
	if err != nil {
		return nil, err
	}

	b := make([]byte, 1)
	_, err = conn.Read(b)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal([]byte{byte(RaftRPC)}, b) {
		return nil, fmt.Errorf("not a raft rpc")
	}
	if s.serverTLSConfig != nil {
		return tls.Server(conn, s.serverTLSConfig), nil
	}
	return conn, nil
}

// 리스너 주소 리턴
func (s *StreamLayer) Addr() net.Addr {
	return s.ln.Addr()
}

// 리스너를 닫는다.
func (s *StreamLayer) Close() error {
	return s.ln.Close()
}

/*
디스커버리 통합
Serf 멤버십이 바뀌면 raft 클러스터도 바뀐다.
서버를 클러스터에 추가할 때마다 Serf 는 멤버가 조인했다는 이벤트를 발행하고, discovery.Membership 은 Join 메서드를 호출한다.
서버가 클러스터를 떠나면 Serf 는 멤버가 떠났다는 이벤트를 발행하고, discovery.Membership 은 Leave 메서드를 호출한다.
DistributedLog 는 Membership 의 핸들러 역할을 하므로, Join 과 Leave 메서드가 raft 를 업데이트하도록 구현해야 한다.
*/
func (l *DistributedLog) Join(id, addr string) error {
	configFuture := l.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == serverID || srv.Address == serverAddr {
			// 이미 조인한 서버
			return nil
		}
		// 기존 서버 삭제
		removeFuture := l.raft.RemoveServer(serverID, 0, 0)
		if err := removeFuture.Error(); err != nil {
			return err
		}
	}
	// 모든 서버를 투표자로 추가한다. AddNotVoter 를 통해 투표하지 못하도록 등록할 수 있다.(읽기만 가능한 서버)
	addFuture := l.raft.AddVoter(serverID, serverAddr, 0, 0)
	if err := addFuture.Error(); err != nil {
		return err
	}
	return nil
}

func (l *DistributedLog) Leave(id string) error {
	// 리더가 아닌 노드를 제거하면 raft 는 ErrNotLeader 에러를 반환한다.
	// 노드가 리더가 아니면 에러로 로그에 기록할 필요가 없다. membership.go 에서 이를 반영한다.
	removeFuture := l.raft.RemoveServer(raft.ServerID(id), 0, 0)
	return removeFuture.Error()
}

func (l *DistributedLog) WaitForLeader(timeout time.Duration) error {
	timeoutc := time.After(timeout)       // 지정 시간 이후에 현재 시간을 채널로 보낸다.
	ticker := time.NewTicker(time.Second) // 1초마다 현재 시간을 채널로 보낸다.
	defer ticker.Stop()

	// 리더를 선출하거나 타임아웃이 걸릴 때까지 대기한다.
	for {
		select {
		case <-timeoutc:
			return fmt.Errorf("timed out")
		case <-ticker.C:
			if l := l.raft.Leader(); l != "" {
				return nil
			}
		}
	}
}

func (l *DistributedLog) Close() error {
	f := l.raft.Shutdown()
	if err := f.Error(); err != nil {
		return err
	}
	return l.log.Close()
}

func (l *DistributedLog) GetServers() ([]*api.Server, error) {
	future := l.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}

	var servers []*api.Server
	for _, server := range future.Configuration().Servers {
		servers = append(servers, &api.Server{
			Id:       string(server.ID),
			RpcAddr:  string(server.Address),
			IsLeader: l.raft.Leader() == server.Address,
		})
	}
	return servers, nil
}
