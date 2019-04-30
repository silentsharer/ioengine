// +build linux
// +build amd64 arm64

package ioengine

import "unsafe"

type iocb struct {
	data   unsafe.Pointer
	key    uint64
	opcode int16
	prio   int16
	fd     uint32
	buf    unsafe.Pointer
	nbytes uint64
	offset int64
	pad1   int64
	flags  uint32
	resfd  uint32
}

type event struct {
	data unsafe.Pointer
	obj  *iocb
	res  int64
	res2 int64
}
