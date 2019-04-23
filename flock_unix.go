// +build linux darwin

package ioengine

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const lockFile = "flock"

// FileLock holds a lock on a hidden file inside.
type FileLock struct {
	path     string
	fd       *os.File
	writable bool
}

// NewFileLock return internal file lock
// example: The file lock format for the file name 'test' is .test-flock
func NewFileLock(path string, writable bool) (*FileLock, error) {
	absLockFile := path
	if filepath.IsAbs(path) {
		absLockFile = fmt.Sprintf("%s/.%s-%s", filepath.Dir(path), filepath.Base(path), lockFile)
	} else {
		absLockFileTmp, err := filepath.Abs(fmt.Sprintf(".%s-%s", path, lockFile))
		if err != nil {
			return nil, err
		}
		absLockFile = absLockFileTmp
	}

	if _, err := os.Stat(absLockFile); err != nil {
		if os.IsNotExist(err) {
			fd, err := os.Create(absLockFile)
			if err != nil {
				return nil, err
			}
			fd.Close()
		}
	}

	return &FileLock{
		path:     absLockFile,
		writable: writable,
	}, nil
}

// FLock a block file lock
func (fl *FileLock) FLock() error {
	fd, err := os.Open(fl.path)
	if err != nil {
		return fmt.Errorf("lock file: %v", err)
	}

	fl.fd = fd
	opts := unix.LOCK_EX
	if !fl.writable {
		opts = unix.LOCK_SH
	}
	if err := unix.Flock(int(fl.fd.Fd()), opts); err != nil {
		return fmt.Errorf("lock file: %v", err)
	}
	return nil
}

// FUnlock file unlock
func (fl *FileLock) FUnlock() error {
	if fl.fd == nil {
		return nil
	}
	if err := unix.Flock(int(fl.fd.Fd()), unix.LOCK_UN); err != nil {
		return fmt.Errorf("unlock file: %v", err)
	}
	return fl.fd.Close()
}

// Release deletes the pid file and auto releases lock on the file
func (fl *FileLock) Release() error {
	if fl.fd != nil {
		fl.fd.Close()
	}
	return os.Remove(fl.path)
}
