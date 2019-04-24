package ioengine

import (
	"fmt"
	"os"
	"testing"
)

const ConcurrentNumber = 100

var fileID int

func NewFileIO() (*FileIO, error) {
	opt := DefaultOptions
	opt.IOEngine = StandardIO
	fileID++
	name := fmt.Sprintf("/tmp/standardio/%d", fileID)
	os.Remove(name)
	return newFileIO(name, opt)
}

func NewFileIOWithAppend() (*FileIO, error) {
	opt := DefaultOptions
	opt.IOEngine = StandardIO
	opt.Flag = os.O_RDWR | os.O_CREATE | os.O_SYNC | os.O_APPEND
	fileID++
	name := fmt.Sprintf("/tmp/standardio/%d", fileID)
	os.Remove(name)
	return newFileIO(name, opt)
}

func TestStandardIOWriteAtv(t *testing.T) {
	fd, err := NewFileIO()
	if err != nil {
		t.Fatalf("Failed to new fileio: %v", err)
	}
	defer fd.Close()

	b := NewBuffers()
	b.Write([]byte("hello")).Write([]byte("world"))

	nw, err := fd.WriteAtv(*b, 2)
	if err != nil {
		t.Fatal(err)
	}
	if nw != 10 {
		t.Fatal("short write")
	}
}

func TestStandardIOAppend(t *testing.T) {
	fd, err := NewFileIO()
	if err != nil {
		t.Fatalf("Failed to new fileio: %v", err)
	}
	defer fd.Close()

	b := NewBuffers()
	b.Write([]byte("hello")).Write([]byte("world"))

	nw, err := fd.Append(*b)
	if err != nil {
		t.Fatal(err)
	}
	if nw != 10 {
		t.Fatal("short write 01")
	}

	nw, err = fd.Append(*b)
	if err != nil {
		t.Fatal(err)
	}
	if nw != 10 {
		t.Fatal("short write 02")
	}

	stat, err := fd.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if stat.Size() != 20 {
		t.Fatal("short write 03")
	}
}

func TestStandardIOComposeWrite(t *testing.T) {
	fd, err := NewFileIO()
	if err != nil {
		t.Fatalf("Failed to new fileio: %v", err)
	}
	defer fd.Close()

	for _, b := range [][]byte{[]byte("12345"), []byte("")} {
		nw, err := fd.Write(b)
		if err != nil {
			t.Fatal(err)
		}
		if nw != len(b) {
			t.Fatal("write: short write")
		}
	}

	b := NewBuffers()
	for i := 0; i < 2; i++ {
		nw, err := fd.Append(*b)
		if err != nil {
			t.Fatal(err)
		}
		if nw != b.Length() {
			t.Fatal("append: short write")
		}
		b.Write([]byte("hello"))
	}

	nw, err := fd.Write([]byte("world"))
	if err != nil {
		t.Fatal(err)
	}
	if nw != 5 {
		t.Fatal("write after append: short write")
	}
}
