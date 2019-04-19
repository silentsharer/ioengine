package ioengine

import (
	"errors"
	"os"
)

// IOMode specifies disk I/O mode, default AIO.
type IOMode int

const (
	// StandardIO indicates that disk I/O using standard buffered I/O
	StandardIO IOMode = iota
	// MMap indicates that disk I/O using memory mapped
	MMap
	// DIO indicates that disk I/O using Direct I/O
	DIO
	// AIO indicates that disk I/O using Async I/O by libaio or io_uring
	AIO
)

// FileLockMode specifies file lock mode, default None.
type FileLockMode int

const (
	// None indicates that open file without file lock
	None FileLockMode = iota
	// ReadWrite indicates that open file with file rwlock
	ReadWrite
	// ReadOnly indicates that open file with file rlock
	ReadOnly
)

// AIOMode specifies aio mode, default Libaio.
type AIOMode int

const (
	// Libaio linux kernel disk async IO solution
	Libaio AIOMode = iota
	// IOUring linux kernel new async IO with v5.1
	IOUring
)

// Options are params for creating IOEngine.
type Options struct {
	IOEngine         IOMode
	Flag             int
	Perm             os.FileMode
	FileLock         FileLockMode
	MmapSize         int
	MmapWritable     bool
	AIO              AIOMode
	AIOMaxQueueDepth int
}

// DefaultOptions is recommended options, you can modify these to suit your needs.
var DefaultOptions = Options{
	IOEngine:         AIO,
	Flag:             os.O_RDWR | os.O_CREATE | os.O_SYNC,
	Perm:             0644,
	FileLock:         None,
	MmapSize:         1<<30 - 1,
	MmapWritable:     true,
	AIO:              Libaio,
	AIOMaxQueueDepth: 256,
}

// File a unified common file operation interface
type File interface {
	Fd() uintptr
	Stat() (os.FileInfo, error)
	Read(b []byte) (int, error)
	ReadAt(b []byte, off int64) (int, error)
	Write(b []byte) (int, error)
	WriteAt(b []byte, off int64) (int, error)
	WriteAtv(bs [][]byte, off int64) (int, error)
	Append(bs [][]byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	Truncate(size int64) error
	FLock() error
	FUnlock() error
	Sync() error
	Close() error
}

// Open opens the named file for reading
func Open(name string, opt Options) (File, error) {
	switch opt.IOEngine {
	case StandardIO:
		return newFileIO(name, opt)
	case MMap:
		return newMemoryMap(name, opt)
	case DIO:
		return newDirectIO(name, opt)
	//case AIO:
	// return newAsyncIO(name, opt)
	default:
		return nil, errors.New("Unsupported IO Engine")
	}
}
