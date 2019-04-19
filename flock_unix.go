// +build linux darwin

package ioengine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const lockFile = "flock"

// FileLock holds a lock on a pid file inside.
type FileLock struct {
	path     string
	fd       *os.File
	writable bool
}

// NewFileLock return file lock
func NewFileLock(path string, writable bool) (*FileLock, error) {
	absLockFile, err := filepath.Abs(filepath.Join(path, lockFile))
	if err != nil {
		return nil, errors.New("can't get absolute path for lock file")
	}

	fd, err := os.Open(absLockFile)
	if err != nil {
		return nil, fmt.Errorf("can't open dir: %v", err)
	}

	flock := &FileLock{
		path:     absLockFile,
		fd:       fd,
		writable: writable,
	}
	return flock, nil
}

// FLock file lock
func (fl *FileLock) FLock() error {
	opts := unix.LOCK_EX | unix.LOCK_NB
	if !fl.writable {
		opts = unix.LOCK_SH | unix.LOCK_NB
	}
	if err := unix.Flock(int(fl.fd.Fd()), opts); err != nil {
		return fmt.Errorf("can't acquire file lock: %v", err)
	}
	return nil
}

// FUnlock file unlock
func (fl *FileLock) FUnlock() error {
	if err := unix.Flock(int(fl.fd.Fd()), unix.LOCK_UN); err != nil {
		return fmt.Errorf("can't unlock file: %v", err)
	}
	return nil
}

// Release deletes the pid file and releases lock on the file
func (fl *FileLock) Release() error {
	err := fl.fd.Close()
	err = os.Remove(fl.path)
	return err
}
