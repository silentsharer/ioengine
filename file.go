package ioengine

import (
	"errors"
	"os"
	"sync"
)

var once sync.Once

// FileIO Standrad I/O mode
type FileIO struct {
	path string
	opt  Options
	*os.File
	*FileLock
}

func newFileIO(name string, opt Options) (*FileIO, error) {
	fd, err := os.OpenFile(name, opt.Flag, opt.Perm)
	if err != nil {
		return nil, err
	}

	fi := &FileIO{path: name, opt: opt, File: fd}

	switch opt.FileLock {
	case None:
	case ReadWrite:
		fi.FileLock, err = NewFileLock(name, true)
		if err != nil {
			fd.Close()
			return nil, err
		}
	case ReadOnly:
		fi.FileLock, err = NewFileLock(name, false)
		if err != nil {
			fd.Close()
			return nil, err
		}
	}

	return fi, nil
}

// FLock a file lock is a recommended lock.
// if file lock not init, we will init once.
func (fi *FileIO) FLock() (err error) {
	if fi.FileLock == nil {
		once.Do(func() {
			if fi.FileLock == nil {
				fi.FileLock, err = NewFileLock(fi.path, true)
			}
		})
	}
	if err != nil {
		return err
	}
	if fi.FileLock == nil {
		return errors.New("Uninitialized lock")
	}
	return fi.FileLock.FLock()
}

// FUnlock file unlock
func (fi *FileIO) FUnlock() error {
	if fi.FileLock == nil {
		return nil
	}
	return fi.FileLock.FUnlock()
}

// Close impl standard File Close method
func (fi *FileIO) Close() error {
	if fi.FileLock == nil {
		return fi.File.Close()
	}
	fi.FileLock.Release()
	return fi.File.Close()
}
