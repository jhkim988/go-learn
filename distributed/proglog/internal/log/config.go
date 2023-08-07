package log

import "github.com/hashicorp/raft"

type Config struct {
	Raft struct {
		raft.Config
		StreamLayer *raft.StreamLayer // raft 추가
		Bootstreap  bool
	}
	Segment struct {
		MaxStoreBytes uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}
}
