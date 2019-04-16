// +build linux

package ioengine

import (
	"os"
)

type AsyncIO struct {
	*os.File
}

func newAsyncIO(name string, opt Options) (*AsyncIO, error) {
	fd, err := OpenFileWithDIO(name, opt.Flag, opt.Perm)
	if err != nil {
		return nil, err
	}
	return &AsyncIO{File: fd}, nil
}
