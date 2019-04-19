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
	var p unsafe.Pointer
	if len(iocbs) > 0 {
		p = unsafe.Pointer(&iocbs[0])
	} else {
		p = unsafe.Pointer(&zero)
	}
	n, _, err := syscall.Syscall(syscall.SYS_IO_SUBMIT, uintptr(ioctx), uintptr(len(iocbs)), uintptr(p))
	if err != nil {
		return 0, os.NewSyscallError("IO_SUBMIT", err)
	}
	return int(n), nil
}

func (ioctx *IOContext) Cancel(iocbs []iocb, events []event) （int, error) {
	var p0, p1 unsafe.Pointer
	if len(iocbs) > 0 {
		p0 = unsafe.Pointer(&iocbs[0])
	} else {
		p0 = unsafe.Pointer(&zero)
	}
	if len(events) > 0 {
		p1 = unsafe.Pointer(&events[0])
	} else {
		p1 = unsafe.Pointer(&zero)
	}
	n, _, err := syscall.Syscall(syscall.SYS_IO_CANCEL, uintptr(ioctx), uintptr(p0), uintptr(p1))
	if err != nil {
		return 0, os.NewSyscallError("IO_CANCEL", err)
	}
	return int(n), nil
}

func (ioctx *IOContext) GetEvents(minnr, nr int, events []event, timeout timespec) (int, error) {
	var p unsafe.Pointer
	if len(events) > 0 {
		p = unsafe.Pointer(&events[0])
	} else {
		p = unsafe.Pointer(&zero)
	}
	n, _, err := syscall.Syscall6(syscall.SYS_IO_GETEVENTS, uintptr(unsafe.Pointer(ioctx)), uintptr(minnr),
		uintptr(nr), uintptr(p), uintptr(unsafe.Pointer(&timeout)), uintptr(0))
	if err != nil {
		return 0, os.NewSyscallError("IO_GETEVENTS", err)
	}
	return int(n), nil
}

func (iocb *iocb) PrepPread(fd int, buf []byte, offset int64) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&buf[0])
	} else {
		p = unsafe.Pointer(&zero)
	}
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPread
	iocb.prio = 0
	iocb.buf = p
	iocb.nbytes = uint64(len(buf))
	iocb.offset = offset
}

func (iocb *iocb) PrepPwrite(fd int, buf []byte, offset int64) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&buf[0])
	} else {
		p = unsafe.Pointer(&zero)
	}
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPwrite
	iocb.prio = 0
	iocb.buf = p
	iocb.nbytes = uint64(len(buf))
	iocb.offset = offset
}

func (iocb *iocb) PrepPreadv(fd int, iovecs []IOvec, offset int64) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&iovecs[0])
	} else {
		p = unsafe.Pointer(&zero)
	}
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPreadv
	iocb.prio = 0
	iocb.buf = p
	iocb.nbytes = uint64(len(iovecs))
	iocb.offset = offset
}

func (iocb *iocb) PrepPwritev(fd int, iovecs []IOvec, offset int64) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&iovecs[0])
	} else {
		p = unsafe.Pointer(&zero)
	}
	iocb.fd = uint32(fd)
	iocb.opcode = IOCmdPwritev
	iocb.prio = 0
	iocb.buf = p
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