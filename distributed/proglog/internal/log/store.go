package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File
	mu sync.Mutex
	buf *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())
	return &store {
		File: f,
		size: size,
		buf: bufio.NewWriter(f),
	}, nil
}

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	// 파일에 직접 쓰지 않고, 버퍼를 거쳐 저장한다.
	// 시스템 콜 횟수를 줄여 성능 개선할 수 있다.
	// 작은 레코드를 많이 호출할 때 효과적
	s.mu.Lock();
	defer s.mu.Unlock();

	pos = s.size

	// 레코드 크기를 쓴다.
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// 레코드를 쓴다.
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth
	s.size += uint64(w)

	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock();
	defer s.mu.Unlock();

	// 데이터가 아직 버퍼에 있는 상황을 대비하여 버퍼를 비운다.
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	// 레코드 크기를 읽는다.
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	// 레코드 크기만큼 읽는다.
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return b, nil
}

// off 부터 len(p) 바이트만큼 읽는다.
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock();
	defer s.mu.Unlock();

	if err := s.buf.Flush(); err != nil {
		return 0, err
	}

	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock();
	defer s.mu.Unlock();

	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
