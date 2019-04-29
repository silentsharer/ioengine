package ioengine

import (
	"fmt"
	"testing"
)

var mmapID int

func NewMemoryMap() (*MemoryMap, error) {
	opt := DefaultOptions
	opt.IOEngine = MMap
	mmapID++
	name := fmt.Sprintf("/tmp/mmap/%d", mmapID)
	// os.Remove(name)
	return newMemoryMap(name, opt)
}

// func TestWrite(t *testing.T) {
// 	mmap, err := NewFileIO()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	mmap.File.Write([]byte("hello"))
// 	mmap.File.Seek(0, 0)
// 	b := make([]byte, 100, 100)
// 	nw, err := mmap.File.ReadAt(b, 0)
// 	t.Log(nw)
// 	t.Log(err)
// 	t.Log(string(b))
// }

func TestMmapWrite(t *testing.T) {
	fd, err := NewMemoryMap()
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	// nw, err := fd.Write([]byte("hello world"))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// if nw != 11 {
	// 	t.Fatal("write: short write")
	// }

	// b := make([]byte, 128, 128)
	// nr, err := fd.Read(b)
	// t.Log(nr)
	// t.Log(err)
	// t.Log(string(b))
}
