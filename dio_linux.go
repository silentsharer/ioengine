// +build linux

package ioengine

import (
	"os"
	"syscall"
)

const (
	// AlignSize size to align the buffer
	AlignSize = 512
	// BlockSize direct IO minimum number of bytes to write
	BlockSize = 4096
)

// OpenFileWithDIO open files with O_DIRECT flag
func OpenFileWithDIO(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, syscall.O_DIRECT|flag, perm)
}

// WriteAtv like linux pwritev, write to the specifies offset and dose not change the file offset.
func (dio *DirectIO) WriteAtv(bs [][]byte, off int64) (int, error) {
	return linuxWriteAtv(dio, bs, off)
}

// Append write data to the end of file.
func (dio *DirectIO) Append(bs [][]byte) (int, error) {
	return genericAppend(dio, bs)
}
