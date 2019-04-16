package ioengine

import (
	"os"
	"syscall"
)

const (
	// AlignSize OSX doesn't need any alignment
	AlignSize = 512
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
