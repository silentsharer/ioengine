// +build linux

package ioengine

import (
	"unsafe"
)

type IOcbCmd int

const (
	IOCmdPread IOcbCmd = iota
	IOCmdPwrite
	IOCmdFSync
	IOCmdFDSync
	IOCmdPoll
	IOCmdNoop
	IOCmdPreadv
	IOCmdPwritev
)

type timespec struct {
	sec  int
	nsec int
}

type IOvec struct {
	Base unsafe.Pointer
	Len uint64
}

type IOContext uint

func (ioctx *IOContext) Setup(maxEvents int) error {
	_, _, err := syscall.Syscall(syscall.SYS_IO_SETUP, uintptr(maxEvents), uintptr(unsafe.Pointer(ioctx)), 0)
	if err != 0 {
		return os.NewSyscallError("IO_SETUP", err)
	}
	return nil
}

func (ioctx *IOContext) Destroy() error {
	_, _, err := syscall.Syscall(syscall.SYS_IO_DESTROY, uintptr(unsafe.Pointer(ioctx)), 0, 0)
	if err != nil {
		return os.NewSyscallError("IO_DESTROY", err)
	}
	return nil
}

func (ioctx *IOContext) Submit(iocbs []iocb) (int, error) {
	n, _, err := syscall.Syscall(syscall.SYS_IO_SUBMIT, uintptr(ioctx), uintptr(len(iocbs)), uintptr(unsafe.Pointer(&iocbs[0])))
	if err != nil {
		return 0, os.NewSyscallError("IO_SUBMIT", err)
	}
	return int(n), nil
}

func (ioctx *IOContext) Cancel(iocbs []iocb, events []event) （int, error) {
	n, _, err := syscall.Syscall(syscall.SYS_IO_CANCEL, uintptr(ioctx), uintptr(unsafe.Pointer(&iocbs[0])), uintptr(unsafe.Pointer(&events[0])))
	if err != nil {
		return 0, os.NewSyscallError("IO_CANCEL", err)
	}
	return int(n), nil
}

func (ioctx *IOContext) GetEvents(minnr, nr int, events []event, timeout timespec) (int, error) {
	n, _, err := syscall.Syscall6(syscall.SYS_IO_GETEVENTS, uintptr(unsafe.Pointer(ioctx)), uintptr(minnr),
		uintptr(nr), uintptr(unsafe.Pointer(&events[0])), uintptr(unsafe.Pointer(&timeout)), uintptr(0))
	if err != nil {
		return 0, os.NewSyscallError("IO_GETEVENTS", err)
	}
	return int(n), nil
}

func (iocb *iocb) PrepPread(fd int, buf []byte, offset int64) {
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPread
	iocb.prio = 0
	iocb.buf = unsafe.Pointer(&buf[0])
	iocb.nbytes = uint64(len(buf))
	iocb.offset = offset
}

func (iocb *iocb) PrepPwrite(fd int, buf []byte, offset int64) {
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPwrite
	iocb.prio = 0
	iocb.buf = unsafe.Pointer(&buf[0])
	iocb.nbytes = uint64(len(buf))
	iocb.offset = offset
}

func (iocb *iocb) PrepPreadv(fd int, iovecs []IOvec, offset int64) {
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPreadv
	iocb.prio = 0
	iocb.buf = unsafe.Pointer(&iovecs[0])
	iocb.nbytes = uint64(len(iovecs))
	iocb.offset = offset
}

func (iocb *iocb) PrepPwritev(fd int, iovecs []IOvec, offset int64) {
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPwritev
	iocb.prio = 0
	iocb.buf = unsafe.Pointer(&iovecs[0])
	iocb.nbytes = uint64(len(iovecs))
	iocb.offset = offset
}

func (iocb *iocb) PrepFSync(fd int) {
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdFSync
	iocb.prio = 0
}

func (iocb *iocb) PrepFDSync(fd int) {
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdFDSync
	iocb.prio = 0
}

func (iocb *iocb) SetEventFd(eventfd int) {
	iocb.flags |= (1<<0)
	iocb.resfd = uint32(eventfd)
}