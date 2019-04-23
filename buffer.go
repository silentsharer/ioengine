package ioengine

import (
	"errors"
	"io"
	"unsafe"
)

// MemAlign like linux posix_memalign.
// block start address must be a multiple of AlignSize.
// block size also must be a multiple of AlignSize.
func MemAlign(blockSize, alignSize uint) ([]byte, error) {
	// make sure blockSize is a multiple of AlignSize.
	if alignSize != 0 && blockSize&(alignSize-1) != 0 {
		return nil, errors.New("invalid argument")
	}
	block := make([]byte, blockSize+alignSize)
	remainder := alignment(block, alignSize)
	var offset uint
	if remainder != 0 {
		offset = alignSize - remainder
	}
	return block[offset : offset+blockSize], nil
}

// alignment returns alignment of the block address in memory with reference to alignSize.
func alignment(block []byte, alignSize uint) uint {
	// if block is nil or length is 0, it will return 0.
	if len(block) < 1 {
		return 0
	}
	// make sure a bit operation mod divisor must be a multiple of 2.
	if alignSize == 0 || alignSize&1 != 0 {
		return 0
	}
	return uint(uintptr(unsafe.Pointer(&block[0])) & uintptr(alignSize-1))
}

// Buffers contains zero or more runs of bytes to write.
// this is applied to readv, writev, preadv, pwritev.
type Buffers [][]byte

// NewBuffers init buffer slice by default cap 128
func NewBuffers() *Buffers {
	buffers := make(Buffers, 0, 128)
	return &buffers
}

func (v *Buffers) Write(b []byte) *Buffers {
	*v = append(*v, b)
	return v
}

func (v *Buffers) Read(b []byte) (n int, err error) {
	for len(b) > 0 && len(*v) > 0 {
		n0 := copy(b, (*v)[0])
		v.consume(int64(n0))
		b = b[n0:]
		n += n0
	}
	if len(*v) == 0 {
		err = io.EOF
	}
	return
}

// WriteTo direct write to writer
func (v *Buffers) WriteTo(w io.Writer) (n int64, err error) {
	for _, b := range *v {
		nb, err := w.Write(b)
		n += int64(nb)
		if err != nil {
			v.consume(n)
			return n, err
		}
	}
	v.consume(n)
	return n, nil
}

func (v *Buffers) consume(n int64) {
	for len(*v) > 0 {
		ln0 := int64(len((*v)[0]))
		if ln0 > n {
			(*v)[0] = (*v)[0][n:]
			return
		}
		n -= ln0
		*v = (*v)[1:]
	}
}
