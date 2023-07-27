package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/jhkim988/proglog/api/v1"
	"google.golang.org/protobuf/proto"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}

	var err error
	
	// store 
	storeFile, err := os.OpenFile(path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}

	// index
	indexFile, err := os.OpenFile(path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")), os.O_RDWR|os.O_CREATE, 0644) 
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}

	// 인덱스의 마지막 항목을 읽어 nextOffset 을 구한다.
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1
	}
	return s, nil
}

func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	cur := s.nextOffset
	record.Offset = cur
	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}

	// store 에 저장
	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}

	// index 에 기록
	if err = s.index.Write(uint32(s.nextOffset - uint64(s.baseOffset)), pos); err != nil {
		return 0, err
	}
	s.nextOffset++
	return cur, nil
}

func (s *segment) Read(off uint64) (*api.Record, error) {
	// offset 에서 baseOffset 을 빼서 segment 에서의 위치로 index 를 읽는다.
	_, pos, err := s.index.Read(int64(off-s.baseOffset))
	if err != nil {
		return nil, err
	}

	// index로부터 읽은 위치로 store 에서의 데이터를 읽는다.
	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}

	// 데이터를 언마샬링한다.
	record := &api.Record{}
	err = proto.Unmarshal(p, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes || s.index.size+entWidth >= s.config.Segment.MaxIndexBytes
}

func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	return nil
}

func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}