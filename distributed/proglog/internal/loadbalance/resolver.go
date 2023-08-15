package loadbalance

import (
	"context"
	"fmt"
	"sync"

	api "github.com/jhkim988/proglog/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

/*
gRPC 리졸버가 GetServers() 엔드포인트를 호출하여 회신받은 정보를 gRPC 에 전달해서 피커가 어느 서버에 요청을 보낼지 판단한다.
리졸버, 피커를 빌더 패턴을 사용한다.
*/

// resolver.Builder, resolver.Resolver 인터페이스 구현
type Resolver struct {
	mu            sync.Mutex
	clientConn    resolver.ClientConn
	resolverConn  *grpc.ClientConn
	serviceConfig *serviceconfig.ParseResult
	logger        *zap.Logger
}

var _ resolver.Builder = (*Resolver)(nil)
var _ resolver.Resolver = (*Resolver)(nil)

const Name = "proglog"

// Build 인터페이스 메서드
// 서버를 찾는 데 필요한 데이터와, 찾아낸 서버 정보로 업데이트할 클라이언트 연결을 받는다.
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.logger = zap.L().Named("resolver")
	r.clientConn = cc
	var dialOpts []grpc.DialOption
	if opts.DialCreds != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(opts.DialCreds))
	}
	r.serviceConfig = r.clientConn.ParseServiceConfig(
		fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, Name),
	)
	var err error
	r.resolverConn, err = grpc.Dial(target.URL.Host, dialOpts...)
	if err != nil {
		return nil, err
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

// Build 인터페이스 메서드
// 리졸버의 Scheme 식별자를 응답 받는다.
func (*Resolver) Scheme() string {
	return Name
}

// 리졸버를 gRPC 에 등록
func init() {
	resolver.Register(&Resolver{})
}

// Resolver 인터페이스 메서드
func (r *Resolver) Close() {
	if err := r.resolverConn.Close(); err != nil {
		r.logger.Error("failed to close conn", zap.Error(err))
	}
}

// Resolver 인터페이스 메서드
// target 에서 정보를 얻고, 서버를 찾아서 클라이언트 연결을 업데이트할 때 호출한다.
// 리졸버가 어떻게 서버를 찾는지에 대해 구현한다.
func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client := api.NewLogClient(r.resolverConn)
	ctx := context.Background()

	// GetServers 요청
	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		r.logger.Error("failed to resolve server", zap.Error(err))
		return
	}

	// GetServers 응답으로 addr 생성, 로드 밸런스가 Address 중에서 서버를 고르게 한다.
	var addrs []resolver.Address
	for _, server := range res.Servers {
		addrs = append(addrs, resolver.Address{
			Addr: server.RpcAddr, // 연결을 위한 서버 주소
			Attributes: attributes.New(
				"is_leader",
				server.IsLeader,
			), // 로드 밸런서에 유용한 데이터를 담은 맵, 어느 서버가 리더이고 팔로워인지 알아내 피커에게 알려준다.
		})
	}

	// 업데이트
	r.clientConn.UpdateState(resolver.State{
		Addresses:     addrs,
		ServiceConfig: r.serviceConfig,
	})
}
