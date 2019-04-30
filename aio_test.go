package ioengine

import (
	"fmt"
	"os"
	"testing"
	"unsafe"
)

var aioID int

func NewAsyncIO() (*AsyncIO, error) {
	opt := DefaultOptions
	opt.IOEngine = AIO
	aioID++
	name := fmt.Sprintf("/tmp/aio/%d", aioID)
	os.Remove(name)

	return newAsyncIO(name, opt)
}

func TestAIOStructure(t *testing.T) {
	var cb iocb
	var evt event

	if unsafe.Sizeof(cb) != 64 {
		t.Fatal(fmt.Sprintf("Invalid iocb structure size: %d != %d", unsafe.Sizeof(cb), 64))
	}
	if unsafe.Sizeof(evt) != 32 {
		t.Fatal(fmt.Sprintf("Invalid event structure size", unsafe.Sizeof(evt), 32))
	}
}

func TestAIONew(t *testing.T) {
	fd, err := NewAsyncIO()
	if err != nil {
		t.Fatal(err)
	}
	if err := fd.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAIOWrite(t *testing.T) {
	fd, err := NewAsyncIO()
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	b, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	copy(b, []byte("hello world"))

	nw, err := fd.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	if nw != len(b) {
		t.Fatal("write: short write")
	}
}

func TestAIOWriteAt(t *testing.T) {
	fd, err := NewAsyncIO()
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	b, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	copy(b, []byte("hello world"))

	nw, err := fd.WriteAt(b, BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	if nw != len(b) {
		t.Fatal("writeAt: short write")
	}

	fi, err := fd.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if fi.Size() != int64(BlockSize+len(b)) {
		t.Fatal("writeAt: invalid file length")
	}
}

func TestAIORead(t *testing.T) {
	fd, err := NewAsyncIO()
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	b, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	copy(b, []byte("hello world"))

	nw, err := fd.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	if nw != len(b) {
		t.Fatal("write: short write")
	}

	rb, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fd.Seek(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	nr, err := fd.Read(rb)
	if err != nil {
		t.Fatal(err)
	}
	if nr != len(b) {
		t.Fatal("read: short read")
	}
}

func TestAIOReadAt(t *testing.T) {
	fd, err := NewAsyncIO()
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	b, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	copy(b, []byte("hello world"))

	nw, err := fd.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	if nw != len(b) {
		t.Fatal("write: short write")
	}

	rb, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}

	nr, err := fd.ReadAt(rb, 0)
	if err != nil {
		t.Fatal(err)
	}
	if nr != len(b) {
		t.Fatal("read: short read")
	}
}

func TestAIOSync(t *testing.T) {
	fd, err := NewAsyncIO()
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	b, err := MemAlign(BlockSize)
	if err != nil {
		t.Fatal(err)
	}
	copy(b, []byte("hello world"))

	nw, err := fd.Write(b)
	if err != nil {
		t.Fatal(err)
	}
	if nw != len(b) {
		t.Fatal("write: short write")
	}

	if err := fd.Sync(); err != nil {
		t.Fatal(err)
	}
}
