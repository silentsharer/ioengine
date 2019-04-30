package ioengine

import (
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
)

// MemoryMap disk IO mode
// page faults and dirty page writes can degrade mmap performance
// we impl ReadAt by mmap, other API impled by standardIO.
type MemoryMap struct {
	path string
	opt  Options
	data []byte
	once sync.Once
	*os.File
	*FileLock
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
		path: name,
		opt:  opt,
		data: data,
		File: fd,
	}

	// runtime.SetFinalizer(mmap, (*mmap).Close())
	return mmap, nil
}

// ReadAt like any io.ReaderAt, clients can execute parallel ReadAt calls.
func (mmap *MemoryMap) ReadAt(b []byte, off int64) (int, error) {
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
	if int64(len(mmap.data)) < off {
		return 0, io.EOF
	}
	n := copy(b, mmap.data[off:])
	if n < len(b) {
		return n, io.EOF
	}

	return n, nil
}

// FLock a file lock is a recommended lock.
// if file lock not init, we will init once
func (mmap *MemoryMap) FLock() (err error) {
	if mmap.FileLock == nil {
		mmap.once.Do(func() {
			if mmap.FileLock == nil {
				mmap.FileLock, err = NewFileLock(mmap.path, true)
			}
		})
	}
	if err != nil {
		return err
	}
	if mmap.FileLock == nil {
		return errors.New("Uninitialized file lock")
	}
	return mmap.FileLock.FLock()
}

// FUnlock file unlock
func (mmap *MemoryMap) FUnlock() error {
	if mmap.FileLock == nil {
		return nil
	}
	return mmap.FileLock.FUnlock()
}

// Close closes the File
func (mmap *MemoryMap) Close() error {
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
