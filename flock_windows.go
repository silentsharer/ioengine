// +build windows

package ioengine

const lockFile = "flock"

type FileLock struct {
}

func NewFileLock(path string, writable bool) (*FileLock, error) {
	return &FileLock{}, nil
}

func (fl *FileLock) FLock() error {
	return nil
}

func (fl *FileLock) FUnlock() error {
	return nil
}

func (fl *FileLock) Release() error {
	return nil
}
