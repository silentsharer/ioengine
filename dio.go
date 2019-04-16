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

// ReadAtv like linux preadv, read from the specifies offset and dose not change the file offset.
func (dio *DirectIO) ReadAtv(off int64, bs ...[]byte) (int, error) {
	return 0, nil
}

// WriteAtv like linux pwritev, write to the specifies offset and dose not change the file offset.
func (dio *DirectIO) WriteAtv(off int64, bs ...[]byte) (int, error) {
	return 0, nil
}

// Append write data to the end of file.
func (dio *DirectIO) Append(bs ...[]byte) (int, error) {
	return 0, nil
}
