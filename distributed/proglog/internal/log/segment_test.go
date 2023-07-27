package log

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	api "github.com/jhkim988/proglog/api/v1"
	"github.com/stretchr/testify/require"
)

func TestSegment(t *testing.T) {
	dir, _ := ioutil.TempDir("", "segment-test")
	defer os.RemoveAll(dir)
	
	want := &api.Record{ Value: []byte("hello world")}

	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entWidth * 3

	s, err := newSegment(dir, 16, c)
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset)
	require.False(t, s.IsMaxed())

	for i:= uint64(0); i < 3; i++ {
		off, err := s.Append(want)
		require.NoError(t, err)
		require.Equal(t, 16+i, off)

		got, err := s.Read(off)
		require.NoError(t, err)
		require.Equal(t, want.Value, got.Value)
	}

	// index 3개까지만 가능하므로 에러
	_, err = s.Append(want)
	require.Equal(t, io.EOF, err)

	c.Segment.MaxStoreBytes = uint64(len(want.Value) * 3)
	c.Segment.MaxIndexBytes = 1024
	s.Close() // 닫아야 Remove test 통과 가능하다!
	s, err = newSegment(dir, 16, c) // 이미 생성된 파일을 읽는데, 스토어 크기를 꽉찬 것으로 설정했다.
	require.NoError(t, err)
	require.True(t, s.IsMaxed())

	err = s.Remove()
	require.NoError(t, err)
	s, err = newSegment(dir, 16, c)
	require.False(t, s.IsMaxed())
}