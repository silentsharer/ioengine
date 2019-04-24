// +build darwin

package ioengine

import (
	"os"
	"syscall"
)

const (
	// AlignSize OSX doesn't need any alignment
	AlignSize = 0
	// BlockSize direct IO minimum number of bytes to write
	BlockSize = 4096
)

// OpenFileWithDIO open files with no cache on darwin.
func OpenFileWithDIO(name string, flag int, perm os.FileMode) (*os.File, error) {
	fd, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	// set no cache
	_, _, er := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd.Fd()), syscall.F_NOCACHE, 1)
	if er != 0 {
		fd.Close()
		return nil, os.NewSyscallError("Fcntl:NoCache", er)
	}

	return fd, nil
}

// WriteAtv simulate writeatv by calling writev serially and dose not change the file offset.
func (dio *DirectIO) WriteAtv(bs [][]byte, off int64) (int, error) {
	return genericWriteAtv(dio, bs, off)
}

// Append write data to the end of file.
// we recommend that open file with O_APPEND
func (dio *DirectIO) Append(bs [][]byte) (int, error) {
	return genericAppend(dio, bs)
}
