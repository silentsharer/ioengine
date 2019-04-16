package ioengine

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const ConcurrentNumber = 100

func NewFileIO() (*FileIO, error) {
	opt := DefaultOptions
	name := "/tmp/fileio"
	return newFileIO(name, opt)
}

func TestSingleProcessConcurrentWrite(t *testing.T) {
	fd, err := NewFileIO()
	if err != nil {
		t.Fatalf("Failed to new fileio: %v", err)
	}
	defer fd.Close()

	fmt.Println(fd.Truncate(-1))

	wg := sync.WaitGroup{}
	fn := func() {
		if _, err := fd.Write([]byte("0123456789")); err != nil {
			t.Fatalf("Failed to write disk: %v", err)
		}
		wg.Done()
	}

	for i := 0; i < ConcurrentNumber; i++ {
		wg.Add(1)
		go fn()
	}
	wg.Wait()
}

func TestSingleProcessConcurrentRead(t *testing.T) {
	fd, err := NewFileIO()
	if err != nil {
		t.Fatalf("Failed to new fileio: %v", err)
	}
	defer fd.Close()

	wg := sync.WaitGroup{}
	fn := func() {
		data := make([]byte, 5, 5)
		if _, err := fd.Read(data); err != nil {
			t.Fatalf("Failed to read disk: %v", err)
		}
		t.Log(string(data))
		wg.Done()
	}

	for i := 0; i < ConcurrentNumber; i++ {
		wg.Add(1)
		go fn()
	}
	wg.Wait()
}

func TestSingleProcessConcurrentReadWrite(t *testing.T) {
	fd, err := NewFileIO()
	if err != nil {
		t.Fatalf("Failed to new fileio: %v", err)
	}
	defer fd.Close()

	readFn := func() {
		data := make([]byte, 5, 5)
		if _, err := fd.Read(data); err != nil {
			t.Fatalf("Failed to read disk: %v", err)
		}
		fmt.Println(string(data))
	}

	writeFn := func() {
		if _, err := fd.Write([]byte("0123456789")); err != nil {
			t.Fatalf("Failed to write disk: %v", err)
		}
	}

	for i := 0; i < ConcurrentNumber; i++ {
		go writeFn()
		go readFn()
	}

	time.Sleep(1000000)
}

func TestMultiProcessConcurrentWrite(t *testing.T) {
	fds := make([]*FileIO, 0, ConcurrentNumber)
	for i := 0; i < ConcurrentNumber; i++ {
		fd, err := NewFileIO()
		if err != nil {
			t.Fatalf("Failed to new fileio: %v", err)
		}
		fds = append(fds, fd)
	}

	wg := sync.WaitGroup{}
	fn := func(fd *FileIO) {
		if _, err := fd.Write([]byte("0123456789")); err != nil {
			t.Fatalf("Failed to write disk: %v", err)
		}
		wg.Done()
	}

	for i := 0; i < ConcurrentNumber; i++ {
		wg.Add(1)
		go fn(fds[i])
	}
	wg.Wait()

	for i := 0; i < ConcurrentNumber; i++ {
		fds[i].Close()
	}
}
