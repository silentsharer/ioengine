// +build linux

package ioengine

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

// WriteAtv like linux pwritev, write to the specifies offset and dose not change the file offset.
func (fi *FileIO) WriteAtv(bs [][]byte, off int64) (int, error) {
	return linuxWriteAtv(fi, bs, off)
}

// Append write data to the end of file.
// we recommend that open file with O_APPEND.
func (fi *FileIO) Append(bs [][]byte) (int, error) {
	return generalAppend(fi, bs)
}

func linuxWriteAtv(fd File, bs [][]byte, off int64) (n int, err error) {
	// read from sysconf(_SC_IOV_MAX)? The Linux default is
	// 1024 and this seems conservative enough for now. Darwin's
	// UIO_MAXIOV also seems to be 1024.
	maxVec := 1024
	var wrote uintptr
	var iovecs []syscall.Iovec

	for len(bs) > 0 {
		iovecs = iovecs[:0]
		for _, chunk := range bs {
			if len(chunk) == 0 {
				continue
			}
			iovecs = append(iovecs, syscall.Iovec{Base: &chunk[0]})
			iovecs[len(iovecs)-1].SetLen(len(chunk))
			if len(iovecs) == maxVec {
				break
			}
		}
		if len(iovecs) == 0 {
			break
		}
		wrote, err = pwritev(int(fd.Fd()), iovecs, off)
		n += int(wrote)
		consume(bs, int64(wrote))
		if err != nil {
			if err.(syscall.Errno) == syscall.EAGAIN {
				continue
			}
			break
		}
		if n == 0 {
			err = io.ErrUnexpectedEOF
			break
		}
	}

	return n, err
}

func pwritev(fd int, iovecs []syscall.Iovec, off int64) (uintptr, error) {
	var p unsafe.Pointer
	if len(iovecs) > 0 {
		p = unsafe.Pointer(&iovecs[0])
	} else {
		p = unsafe.Pointer(&zero)
	}

	n, _, err := syscall.Syscall6(syscall.SYS_PWRITEV, uintptr(fd), uintptr(p), uintptr(len(iovecs)), uintptr(off), 0, 0)
	if err != nil {
		return 0, os.NewSyscallError("PWRITEV", err)
	}

	return n, nil
}

// consume removes data from a slice of byte slices, for writev.
func consume(v [][]byte, n int64) {
	for len(v) > 0 {
		ln0 := int64(len(v[0]))
		if ln0 > n {
			v[0] = v[0][n:]
			return
		}
		n -= ln0
		v = v[1:]
	}
}
