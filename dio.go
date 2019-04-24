package ioengine

import (
	"os"
)

// DirectIO dio mode
type DirectIO struct {
	*os.File
}

func newDirectIO(name string, opt Options) (*DirectIO, error) {
	fd, err := OpenFileWithDIO(name, opt.Flag, opt.Perm)
	if err != nil {
		return nil, err
	}

	return &DirectIO{File: fd}, nil
}

func (dio *DirectIO) FLock() error {
	return nil
}

func (dio *DirectIO) FUnlock() error {
	return nil
}

func (dio *DirectIO) Close() error {
	return nil
}

func (dio *DirectIO) Option() Options {
	return Options{}
}
