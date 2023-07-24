package server

import (
	"fmt"
	"sync"
)

type Log struct {
	mu sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

func (c *Log) Append(record Record) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	record.Offset = uint64(len(c.records))
	c.records = append(c.records, record)
	return record.Offset, nil
}

func (c *Log) Read(offset uint64) (Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if offset >= uint64(len(c.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return c.records[offset], nil
}

var ErrOffsetNotFound = fmt.Errorf("offset not found")

type Record struct {
	Value []byte `json:"value"` // 주의: []byte 타입이므로 요청 보낼 때 value 를 base64로 인코딩한다.
	Offset uint64 `json:"offset"`
}

