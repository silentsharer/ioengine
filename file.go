package ioengine

import "os"

// FileIO Standrad I/O mode
type FileIO struct {
	opt Options
	*os.File
}

func newFileIO(name string, opt Options) (*FileIO, error) {
	fd, err := os.OpenFile(name, opt.Flag, opt.Perm)
	if err != nil {
		return nil, err
	}
	return &FileIO{File: fd, opt: opt}, nil
}
