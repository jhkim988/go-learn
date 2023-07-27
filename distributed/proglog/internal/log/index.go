package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

// 레코드 오프셋 uint32
// 스토어 파일에서의 위치 int64
var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

// 인덱스 파일
// 파일과 메모리 맵 파일
// size 는 인덱스의 크기
type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	// 파일 크기를 idx.size 에 저장한다.
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())

	// 파일 크기를 변경한다.
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	// 메모리맵 생성하여 저장한다.
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}

	return idx, nil
}

func (i *index) Close() error {
	// 메모리맵과 파일을 동기화시킨다.
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}

	// Unmap 해야 Truncate 가능하다!
	if err := i.mmap.UnsafeUnmap(); err != nil {
		return err
	}

	// 파일의 현재 내용을 안정적인 저장소에 커밋한다. (메모리 -> 디스크)
	if err := i.file.Sync(); err != nil {
		return err
	}
	// 실제 데이터가 있는만큼 잘라낸다.(truncate)
	// 레코드를 로그의 어디에 추가할지 오프셋을 알아야 하는데, 마지막 항목의 인덱스를 찾아보면 다음 레코드의 오프셋을 알 수 있다.(인덱스의 마지막 12바이트)
	// 메모리맵 파일을 사용하기 위해 파일크기를 최대로 늘리면 이 방법을 사용할 수 없다. (뒤에 빈공간이 있으면 안됨)
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}

	return i.file.Close()
}
// 비정상종료 - 파일을 잘라내는 도중 전원이 끊길 수 있다.
// 이런 상황을 고려하여 온전성검사를 할 수 있다.

// offset 을 매개변수로 받아, 해당하는 레코드의 저장 파일 내 위치를 리턴
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	if in == -1 {
		// 인덱스의 마지막 엔트리 읽음
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}

	pos = uint64(out) * entWidth
	if (i.size < pos+entWidth) {
		return 0, 0, io.EOF
	}

	out = enc.Uint32(i.mmap[pos: pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth: pos+entWidth])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += uint64(entWidth)
	return nil;
}

func (i*index) Name() string {
	return i.file.Name()
}
