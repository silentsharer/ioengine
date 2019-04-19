package ioengine

import (
	"io"
	"os"
	"syscall"
)

// Single-word zero for use when we need a valid pointer to 0 bytes.
var zero uintptr

func genericWriteAtv(fd File, bs [][]byte, off int64) (n int, err error) {
	nOffset := off
	nw := 0

	for _, b := range bs {
		nw, err = fd.WriteAt(b, nOffset)
		n += nw
		nOffset += int64(nw)
		if err != nil {
			if err.(syscall.Errno) == syscall.EAGAIN {
				continue
			}
			break
		}
		if nw == 0 {
			err = io.ErrUnexpectedEOF
			break
		}
	}

	return n, err
}

func genericAppend(fd File, bs [][]byte) (int, error) {
	stat, err := fd.Stat()
	if err != nil {
		return 0, err
	}

	// open file with O_APPEND not need to flock
	if stat.Mode & os.O_APPEND {
		return fd.WriteAtv(bs, stat.Size())
	}

	if err := fd.FLock(); err != nil {
		return err
	}
	err = fd.WriteAtv(bs, stat.Size())
	return fd.FUnlock()
}
