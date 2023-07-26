package log

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "index_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	c := Config{}
	c.Segment.MaxIndexBytes = 1024
	idx, err := newIndex(f, c)
	require.NoError(t, err)

	_, _, err = idx.Read(-1)
	require.Error(t, err) // 인덱스를 생성하고 아무것도 없으므로 에러
	require.Equal(t, f.Name(), idx.Name())

	entries := []struct {
		Off uint32
		Pos uint64
	} {
		{Off: 0, Pos: 0},
		{Off: 1, Pos: 10},
	}
	for _, want := range entries {
		err = idx.Write(want.Off, want.Pos)
		require.NoError(t, err)

		_, pos, err := idx.Read(int64(want.Off))
		require.NoError(t, err)
		require.Equal(t, want.Pos, pos)
	}
	
	// 존재하는 항목의 범위를 넘어서서 읽기를 시도
	_, _, err = idx.Read(int64(len(entries)))
	require.Error(t, err)

	// index 가 닫혔다가 다시 열릴 때(서비스 재시작) 초기값 테스트
	err = idx.Close()
	require.NoError(t, err)
	f, _= os.OpenFile(f.Name(), os.O_RDWR, 0600)
	idx, err = newIndex(f, c)
	require.NoError(t, err)

	off, pos, err := idx.Read(-1)
	require.NoError(t, err)
	require.Equal(t, entries[1].Off, off)
	require.Equal(t, entries[1].Pos, pos)
}