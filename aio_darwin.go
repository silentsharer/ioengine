// +build !linux

package ioengine

import (
	"errors"
	"os"
)

type AsyncIO struct {
	*os.File
}

func newAsyncIO(name string, opt Options) (*AsyncIO, error) {
	return nil, errors.New("Please use AIO on linux")
}

func (aio *AsyncIO) WriteAtv(bs [][]byte, off int64) (int, error) {
	return 0, nil
}

func (aio *AsyncIO) Append(bs [][]byte) (int, error) {
	return 0, nil
}

func (aio *AsyncIO) FLock() error {
	return nil
}

func (aio *AsyncIO) FUnlock() error {
	return nil
}

func (aio *AsyncIO) Option() Options {
	return Options{}
}
