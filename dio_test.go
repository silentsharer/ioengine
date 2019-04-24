package ioengine

import (
	"fmt"
	"os"
	"testing"
)

var DIOID int

func NewDirectIO() (*DirectIO, error) {
	opt := DefaultOptions
	opt.IOEngine = DIO
	DIOID++
	name := fmt.Sprintf("/tmp/directio/%d", DIOID)
	os.Remove(name)
	return newDirectIO(name, opt)
}

func TestDirectIOWrite(t *testing.T) {
	fd, err := NewDirectIO()
	if err != nil {
		t.Fatalf("Failed to new fileio: %v", err)
	}
	defer fd.Close()

	b0, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	copy(b0, []byte("hello world"))

	nw, err := fd.Write(b0)
	if err != nil {
		t.Fatal(err)
	}
	if nw != BlockSize {
		t.Fatal("write: short write")
	}

	b1, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	copy(b1, []byte("direct IO"))

	b := NewBuffers()
	b.Write(b0).Write(b1)

	nw, err = fd.WriteAtv(*b, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nw != 2*BlockSize {
		t.Fatal("buffers: short write")
	}

	nw, err = fd.Append(*b)
	if err != nil {
		t.Fatal(err)
	}
	if nw != 2*BlockSize {
		t.Fatal("append: short write")
	}
}
