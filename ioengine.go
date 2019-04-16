package ioengine

import (
	"errors"
	"os"
)

// IOEngine specifies disk I/O mode
type IOEngine int

const (
	// StandardIO indicates that disk I/O using standard buffered I/O
	StandardIO IOEngine = iota
	// MMap indicates that disk I/O using memory mapped
	MMap
	// DIO indicates that disk I/O using Direct I/O
	DIO
	// AIO indicates that disk I/O using Async I/O by Libaio
	AIO
)

// Options are params for creating IOEngine.
type Options struct {
	FileIOEngine     IOEngine
	Flag             int
	Perm             os.FileMode
	MmapSize         int
	MmapWritable     bool
	AIOMaxQueueDepth int
}

// DefaultOptions is recommended options, you can modify these to suit your needs.
var DefaultOptions = Options{
	Flag:             os.O_RDWR | os.O_CREATE | os.O_SYNC,
	Perm:             0644,
	FileIOEngine:     StandardIO,
	MmapSize:         1<<30 - 1,
	AIOMaxQueueDepth: 256,
}

// File a unified common file operation interface
type File interface {
	Fd() uintptr
	Stat() (os.FileInfo, error)
	Read(b []byte) (int, error)
	ReadAt(b []byte, off int64) (int, error)
	ReadAtv(off int64, bs ...[]byte) (int, error)
	Write(b []byte) (int, error)
	WriteAt(b []byte, off int64) (int, error)
	WriteAtv(off int64, bs ...[]byte) (int, error)
	Append(bs ...[]byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	Truncate(size int64) error
	Sync() error
	Close() error
}

// Open opens the named file for reading
func Open(name string, opt Options) (File, error) {
	switch opt.FileIOEngine {
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
