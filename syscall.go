package ioengine

import (
	"os"
)

// Single-word zero for use when we need a valid pointer to 0 bytes.
var zero uintptr

// simulate writeatv by calling writeat serially and dose not change the file offset.
func genericWriteAtv(fd File, bs [][]byte, off int64) (n int, err error) {
	nOffset := off
	nw := 0

	for _, b := range bs {
		nw, err = fd.WriteAt(b, nOffset)
		n += nw
		nOffset += int64(nw)
		if err != nil {
			break
		}
	}

	return n, err
}

// simulate writev by calling write serially and it will change the file offset.
func genericWritev(fd File, bs [][]byte) (n int, err error) {
	nw := 0

	for _, b := range bs {
		nw, err = fd.Write(b)
		n += nw
		if err != nil {
			break
		}
	}

	return n, err
}

func genericAppend(fd File, bs [][]byte) (int, error) {
	opt := fd.Option()

	// open file with O_APPEND not need to flock
	if (opt.Flag & os.O_APPEND) > 0 {
		return genericWritev(fd, bs)
	}

	// acquire file size to append write
	size, err := fd.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}
	// Because use writeat to simulate an append write
	// it doesn't change the file offset, to keep append semantic
	// so that make sure file offset is the file end.
	defer fd.Seek(0, os.SEEK_END)

	return fd.WriteAtv(bs, size)
}
