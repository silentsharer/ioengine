package ioengine

import (
	"errors"
	"os"
)

// IOMode specifies disk I/O mode, default StandardIO.
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
	// IOEngine io mode
	IOEngine IOMode

	// Flag the file open mode
	Flag int

	// Perm the file perm
	Perm os.FileMode

	// FileLock file lock mode, default none
	FileLock FileLockMode

	// MmapSize mmap file size in memory
	MmapSize int

	// MmapWritable whether to allow mmap write
	// if true, it will be use mmap write instead of standardIO write, not implemented yet.
	MmapWritable bool

	// AIO async IO mode, defaul libaio, the io_uring isn't implemented yet.
	AIO AIOMode

	// AIOQueueDepth libaio max events, it's also use to control client IO number.
	AIOQueueDepth int

	// AIOTimeout unit ms, libaio timeout, 0 means no timeout.
	AIOTimeout int
}

// DefaultOptions is recommended options, you can modify these to suit your needs.
var DefaultOptions = Options{
	IOEngine:      StandardIO,
	Flag:          os.O_RDWR | os.O_CREATE | os.O_SYNC,
	Perm:          0644,
	FileLock:      None,
	MmapSize:      1<<30 - 1,
	MmapWritable:  false,
	AIO:           Libaio,
	AIOQueueDepth: 1024,
	AIOTimeout:    0,
}

// File a unified common file operation interface
type File interface {
	// Fd returns the Unix fd or Windows handle referencing the open file.
	// The fd is valid only until f.Close is called or f is garbage collected.
	Fd() uintptr

	// Stat returns the FileInfo structure describing file.
	// The MMap mode returns the native file state instead of the memory slice.
	Stat() (os.FileInfo, error)

	// Read reads up to len(b) bytes from the File.
	// It returns the number of bytes read and any error encountered.
	// At end of file, Read returns io.EOF.
	Read(b []byte) (int, error)

	// ReadAt reads len(b) bytes from the File starting at byte offset off.
	// It returns the number of bytes read and the error, if any.
	// ReadAt always returns a non-nil error when n < len(b).
	// At end of file, that error is io.EOF.
	ReadAt(b []byte, off int64) (int, error)

	// Write writes len(b) bytes to the File.
	// It returns the number of bytes written and an error, if any.
	// Write returns a non-nil error when n != len(b).
	Write(b []byte) (int, error)

	// WriteAt writes len(b) bytes to the File starting at byte offset off.
	// It returns the number of bytes written and an error, if any.
	// WriteAt returns a non-nil error when n != len(b).
	WriteAt(b []byte, off int64) (int, error)

	// WriteAtv write multiple discrete discontinuous mem block
	// on AIO mode, it's impled by pwritev syscall
	// on other mode, it's impled by multi call pwrite syscall
	WriteAtv(bs [][]byte, off int64) (int, error)

	// Append write data at the end of file
	// We do not guarantee atomicity of concurrent append writes.
	Append(bs [][]byte) (int, error)

	// Seek sets the offset for the next Read or Write on file to offset, interpreted
	// according to whence: 0 means relative to the origin of the file, 1 means
	// relative to the current offset, and 2 means relative to the end.
	// It returns the new offset and an error, if any.
	// The behavior of Seek on a file opened with O_APPEND is not specified.
	Seek(offset int64, whence int) (int64, error)

	// Truncate changes the size of the file.
	// It does not change the I/O offset.
	// If there is an error, it will be of type *PathError.
	Truncate(size int64) error

	// FLock the lock is suggested and exclusive
	FLock() error

	// FUnlock unlock the file lock
	// it will be atomic release when file close.
	FUnlock() error

	// Sync commits the current contents of the file to stable storage.
	// Typically, this means flushing the file system's in-memory copy
	// of recently written data to disk.
	Sync() error

	// Close closes the File, rendering it unusable for I/O.
	Close() error

	// Option return IO engine options
	Option() Options
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
	case AIO:
		return newAsyncIO(name, opt)
	default:
		return nil, errors.New("Unsupported IO Engine")
	}
}
