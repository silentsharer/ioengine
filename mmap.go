package ioengine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

// MemoryMap disk IO mode
type MemoryMap struct {
	path string
	opt  Options
	// simulate linux file max size
	data []byte
	// simulate linux file offset
	offset int
	// simulate linux file end
	end int
	*os.File
	sync.RWMutex
}

func newMemoryMap(name string, opt Options) (*MemoryMap, error) {
	fd, err := os.OpenFile(name, opt.Flag, opt.Perm)
	if err != nil {
		return nil, err
	}

	data, err := Mmap(fd, 0, opt.MmapSize, opt.MmapWritable)
	if err != nil {
		return nil, err
	}
	if err := Madvise(data); err != nil {
		return nil, err
	}

	mmap := &MemoryMap{
		path:   name,
		opt:    opt,
		data:   data,
		offset: 0,
		end:    0,
		File:   fd,
	}

	// recalc mmap end
	stat, err := fd.Stat()
	if err != nil {
		return nil, err
	}
	if int64(opt.MmapSize) < stat.Size() {
		mmap.end = opt.MmapSize
	} else {
		mmap.end = int(stat.Size())
	}

	// runtime.SetFinalizer(mmap, (*mmap).Close())
	return mmap, nil
}

// Fd wraps standard I/O Fd
func (mmap *MemoryMap) Fd() uintptr {
	return mmap.File.Fd()
}

// Stat wraps standard I/O Stat
// maybe should return the mmaped slice's FileInfo not the raw file's FileInfo
func (mmap *MemoryMap) Stat() (os.FileInfo, error) {
	return mmap.File.Stat()
}

// Read like any io.Read, Shares the same file offset.
func (mmap *MemoryMap) Read(b []byte) (int, error) {
	mmap.RLock()
	defer mmap.RUnlock()

	// If the caller wanted a zero byte read, return immediately.
	if len(b) == 0 {
		return 0, nil
	}
	if mmap.data == nil {
		return 0, errors.New("mmap: closed")
	}
	if mmap.offset >= mmap.end {
		return 0, io.EOF
	}
	n := copy(b, mmap.data[mmap.offset:mmap.end])
	mmap.offset += n
	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// ReadAt like any io.ReaderAt, clients can execute parallel ReadAt calls.
func (mmap *MemoryMap) ReadAt(b []byte, off int64) (int, error) {
	mmap.RLock()
	defer mmap.RUnlock()

	// If the caller wanted a zero byte read, return immediately.
	if len(b) == 0 {
		return 0, nil
	}
	if mmap.data == nil {
		return 0, errors.New("mmap: closed")
	}
	if off < 0 {
		return 0, errors.New("negative offset")
	}
	if int64(mmap.end) < off {
		return 0, io.EOF
	}
	n := copy(b, mmap.data[off:mmap.end])
	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// Write like any io.Write, Shares the same file offset.
// If the mmaped file size is reached, it will return EOF.
func (mmap *MemoryMap) Write(b []byte) (int, error) {
	mmap.Lock()
	defer mmap.Unlock()

	if len(b) == 0 {
		return 0, nil
	}
	if mmap.data == nil {
		return 0, errors.New("mmap: closed")
	}
	if mmap.offset >= len(mmap.data) {
		return 0, io.EOF
	}

	n := copy(mmap.data[mmap.offset:], b)
	mmap.offset += n
	mmap.reCalcEnd(mmap.offset)
	if n < len(b) {
		return n, io.ErrShortWrite
	}

	return n, nil
}

// WriteAt like any io.WriteAt, clients can execute parallel WriteAt calls without panic.
func (mmap *MemoryMap) WriteAt(b []byte, off int64) (int, error) {
	mmap.Lock()
	defer mmap.Unlock()

	if mmap.data == nil {
		return 0, errors.New("mmap: closed")
	}
	if off < 0 || int64(len(mmap.data)) < off {
		return 0, fmt.Errorf("mmap: invalid WriteAt offset %d", off)
	}
	n := copy(mmap.data[off:], b)
	mmap.reCalcEnd(int(off) + n)
	if n < len(b) {
		return n, io.ErrShortWrite
	}

	return n, nil
}

func (mmap *MemoryMap) WriteAtv(bs [][]byte, off int64) (int, error) {
	mmap.Lock()
	defer mmap.Unlock()

	if mmap.data == nil {
		return 0, errors.New("mmap: closed")
	}
	if off < 0 || int64(len(mmap.data)) < off {
		return 0, fmt.Errorf("mmap: invalid WriteAt offset %d", off)
	}

	var n, nw, nOffset int
	nOffset = int(off)
	for _, b := range bs {
		nw = copy(mmap.data[nOffset:], b)
		mmap.reCalcEnd(int(nOffset) + nw)
		nOffset += nw
		n += nw
		if nw < len(b) {
			return n, io.ErrShortWrite
		}
	}

	return n, nil
}

// Append write data to the end of file.
func (mmap *MemoryMap) Append(bs [][]byte) (int, error) {
	mmap.Lock()
	defer mmap.Unlock()

	if mmap.data == nil {
		return 0, errors.New("mmap: closed")
	}

	return mmap.WriteAtv(bs, int64(mmap.end))
}

// Seek like any io.Seek, if the new offset is greater than mmaped file size, it will return error.
func (mmap *MemoryMap) Seek(offset int64, whence int) (int64, error) {
	mmap.Lock()
	defer mmap.Unlock()

	var nOffset int64
	switch whence {
	case 0:
		nOffset = offset
	case 1:
		nOffset = int64(mmap.offset) + offset
	case 2:
		nOffset = int64(mmap.end) + offset
	}

	if nOffset < 0 || nOffset > int64(len(mmap.data)) {
		return 0, errors.New("invalid argument")
	}
	mmap.offset = int(nOffset)

	return nOffset, nil
}

// Truncate changes the size of file, it does not change I/O offset
// if the truncated size is greater than mmaped file size, it will return error.
func (mmap *MemoryMap) Truncate(size int64) error {
	mmap.Lock()
	defer mmap.Unlock()

	if size < 0 || size > int64(len(mmap.data)) {
		return errors.New("invalid argument")
	}

	for i := int(size); i < mmap.end; i++ {
		mmap.data[i] = 0
	}

	mmap.end = int(size)
	return nil
}

// Sync wraps mmap sync
func (mmap *MemoryMap) Sync() error {
	mmap.Lock()
	defer mmap.Unlock()
	return Sync(mmap.data)
}

func (mmap *MemoryMap) FLock() (err error) {
	return nil
}

func (mmap *MemoryMap) FUnlock() error {
	return nil
}

// Close closes the File
func (mmap *MemoryMap) Close() error {
	mmap.Lock()
	defer mmap.Unlock()

	if mmap.File != nil {
		mmap.File.Close()
	}
	if mmap.data == nil {
		return nil
	}

	data := mmap.data
	mmap.data = nil
	runtime.SetFinalizer(mmap, nil)
	return Munmap(data)
}

// Option return file options
func (mmap *MemoryMap) Option() Options {
	return mmap.opt
}

func (mmap *MemoryMap) reCalcEnd(off int) {
	if off > mmap.end {
		mmap.end = off
	}
}
