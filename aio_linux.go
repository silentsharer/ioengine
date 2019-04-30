// +build linux

package ioengine

import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

const (
	defaultQueueDepth int = 1024
)

var (
	ErrNotInit           = errors.New("Not initialized")
	ErrWaitAllFailed     = errors.New("Failed to wait for all requests to complete")
	ErrNilCallback       = errors.New("The kernel returned a nil callback iocb structure")
	ErrUntrackedEventKey = errors.New("The kernel returned an event key we weren't tracking")
	ErrInvalidEventPtr   = errors.New("The kernel returned an invalid callback event pointer")
	ErrReqIDNotFound     = errors.New("The requestID not found")
	ErrNotDone           = errors.New("Request not finished")
)

// RequestID aio submit request id
type RequestID uint

type runningEvent struct {
	data  []byte
	wrote uint
	iocb  *iocb
	reqID RequestID
}

type requestState struct {
	iocb  *iocb
	done  bool
	err   error
	bytes int64
}

// AsyncIO async IO
// maybe we can implement a simplified posix file system?
type AsyncIO struct {
	opt    Options

	// fd raw file descriptor 
	fd     *os.File

	// ioctx the value after initialization isn't 0
	ioctx  IOContext

	// iocbs it's used to do IO
	iocbs  []*iocb

	// events it's used to capture completed IO event
	events []event

	// offset read and write file offset
	offset int64

	// end the end of file
	end   int64

	// reqID every read or write auto incre id
	reqID RequestID

	// running the pool of commited IO
	// map structure: map[*iocb]*runningEvent
	running ConcurrentMap

	// available the pool of available IO, it's max vale is opt.AIOQueueDepth
	// if is no available iocb, the IO(read, write) will be block until it has.
	// map structure: map[*iocb]bool
	available ConcurrentMap

	// request pool record every IO stat
	// map structure: map[RequestID]*requestState
	request ConcurrentMap

	// close wait all running IO completed and finish wait
	close chan struct{}

	// trigger wait all running IO completed
	trigger chan struct{}

	sync.RWLock
}

func newAsyncIO(name string, opt Options) (*AsyncIO, error) {
	fd, err := OpenFileWithDIO(name, opt.Flag, opt.Perm)
	if err != nil {
		return nil, err
	}

	// verify aio queue depth
	if opt.AIOQueueDepth <= 0 || opt.AIOQueueDepth > defaultQueueDepth {
		opt.AIOQueueDepth = defaultQueueDepth
	}

	// fetch the end of file
	stat, err := fd.Stat()
	if err != nil {
		fd.Close()
		return nil, err
	}
	end := stat.Size()

	ioctx, err := NewIOContext()
	if err != nil {
		fd.Close()
		return nil, err
	}

	// init iocbs and available pool
	available := NewConcurrentMap()
	iocbs := make([]*iocb, opt.AIOQueueDepth)
	for i := range iocbs {
		iocbs[i] = NewIocb(uint32(fd.Fd()))
		available.Set(pointer2string(unsafe.Pointer(iocbs[i])), iocbs[i])
	}

	aio := &AsyncIO{
		File:      fd,
		ioctx:     ioctx,
		iocbs:     iocbs,
		events:    make([]event, opt.AIOQueueDepth),
		offset:    0,
		end:       end,
		reqID:     1,
		available: available,
		running:   NewConcurrentMap(),
		request:   NewConcurrentMap(),
	}

	// start a goroutine loop to fetch completed IO
	go aio.wait()

	return aio, nil
}

// Close will wait for all submitted IO to completed.
func (aio *AsyncIO) Close() error {
	if aio.ioctx == 0 {
		return ErrNotInit
	}

	// send to signal to stop wait
	aio.close <- struct{}
	<-aio.close

	// destroy async IO context
	aio.ioctx.Destroy()
	aio.ioctx = 0

	// close file descriptor
	if err := aio.fd.Close(); err != nil {
		return err
	}

	return nil
}

func (aio *AsyncIO) wait() {
	for {
		select {
		case <-close:
			wait()
			close <- struct{}
			return
		case <-trigger:
			wait()
			trigger <- struct{}
		default:
			wait()
		}
	}
}

func (aio *AsyncIO) waitEvents() error {
	numRunningIO := running.Count()
	if numRunningIO == 0 {
		return nil
	}

	t := timeout{
		sec:  aio.opt.AIOTimeout / 1000,
		nsec: (aio.opt.AIOTimeout % 1000) * 1000 * 1000,
	}

	// wait for at least one running IO to complete.
	n, err := aio.ioctx.GetEvents(1, uintptr(numRunningIO), aio.events, t)
	if err != nil {
		return err
	}
	if n == 0 || n > numRunningIO {
		return ErrWaitAllFailed
	}

	for i := 0; i < n; i++ {
		if e := aio.verifyEvent(aio.events[i]); e != nil {
			err = e
		}
	}

	return err
}

// verifyEvent checks that a retuned event is for a valid request
func (aio *AsyncIO) verifyEvent(evt event) error {
	if evt.obj == nil {
		return ErrNilCallback
	}
	revt, ok := aio.running.Get(pointer2string(unsafe.Pointer(evt.obj)))
	if !ok {
		return ErrUntrackedEventKey
	}
	if revt.obj != evt.obj {
		return ErrInvalidEventPtr
	}
	// an error occured with this event, remove the running event and set error code.
	if evt.res < 0 {
		return aio.freeEvent(revt, evt.iocb, lookupErrNo(evt.res))
	}
	//we have an active event returned and its one we are tracking
	//ensure it wrote our entire buffer, res is > 0 at this point
	if evt.res > 0 && uint(len(revt.data)) != uint(evt.res)+revt.wrote {
		revt.wrote += uint(evt.res)
		if err := aio.resubmit(); err != nil {
			return err
		}
		return nil
	}
	revt.wrote += uint(evt.res)

	return aio.freeEvent(revt, evt.iocb, nil)
}

// resubmit puts a request back into the kernel
// this is done when a partial read or write occurs
func (aio *AsyncIO) resubmit(re *runningEvent) error {
	// double check we are not about to roll outside our buffer
	if re.wrote >= uint(len(re.data)) {
		return nil
	}

	nBuf := re.data[re.wrote:]
	nOffset := re.iocb.offset + int64(re.wrote)
	switch re.iocb.OpCode() {
	case IOCmdPread:
		re.iocb.PrepPread(nBuf, nOffset)
	case IOCmdPwrite:
		re.iocb.PrepPwrite(nBuf, nOffset)
	case IOCmdPreadv:
	case IOCmdPwritev:
	}

	return aio.ioctx.Submit([]iocb{*re.iocb})
}

// freeEvent removes an running event and return its iocb to the available pool
func (aio *AsyncIO) freeEvent(re *runningEvent, iocb *iocb, err error) error {
	// help gc free memory early
	re.data = nil

	// remove the iocb from running pool
	aio.running.Remove(pointer2string(unsafe.Pointer(re.iocb)))

	// put the iocb back into the available pool
	aio.available.Set(pointer2string(unsafe.Pointer(re.iocb)), re.iocb)

	// update the stat in request pool
	r, ok := aio.request.Get(int2string(int64(re.reqID)))
	if !ok {
		return ErrReqIDNotFound
	}
	r.done = true
	r.bytes = int64(re.wrote)
	if err != nil {
		r.err = err
	}

	return nil
}

// getNextReady will retrieve the next available iocb for use
// if no iocb are available, it blocks and waits for one.
func (aio *AsyncIO) getNextReady() *iocb {
	for {
		_, v, has := aio.available.RandomPop()
		if has {
			return (*iocb)(unsafe.Pointer(v))
		}
	}
}

// waitAll will block until all submitted io are done
func (aio *AsyncIO) waitAll() {
	trigger <- struct{}
	<-trigger
}

// WaitFor will block until the given RequestId is done
func (aio *AsyncIO) WaitFor(id RequestID) (int, error) {
	for {
		// check if its ready
		done, err := aio.IsDone(id)
		if err != nil {
			return 0, err
		}
		if done {
			break
		}
	}

	return aio.ack(id)
}

func (aio *AsyncIO) IsDone(id RequestID) (bool, error) {
	v, ok := aio.request.Get(int2string(int64(id)))
	if !ok {
		return false, ErrReqIDNotFound
	}
	return (*requestState)(v).done, nil
}

// Ack acknowledges that we have accepted a finished result ID
// if the request is not done, an error is returned
func (aio *AsyncIO) ack(id RequestID) (int, error) {
	v, ok := aio.request.Get(int2string(int64(id)))
	if !ok {
		return 0, ErrReqIDNotFound
	}
	stat := (*requestState)(v)
	if stat.done {
		aio.request.Remove(int2string(int64(id)))
		return stat.bytes, stat.err
	}
	return 0, ErrNotDone
}

func (aio *AsyncIO) reCalcEnd(offset int64) {
	if offset > aio.end {
		aio.end = offset
	}
}

// Fd wraps *os.File Fd to impl File Fd func
func (aio *AsyncIO) Fd() uintptr {
	return aio.fd.Fd()
}

// Stat wraps *os.File Stat to impl File Stat func
func (aio *AsyncIO) Stat() (os.FileInfo, error) {
	return aio.fd.Stat()
}

// Write simulate write by writeAt, it is a async IO.
// the buffer cannot change before the write completes.
func (aio *AsyncIO) Write(b []byte) (int, error) {
	nw, err := aio.WriteAt(b, aio.offset)
	aio.offset += nw
	return nw, err
}

func (aio *AsyncIO) WriteAt(b []byte, offset int64) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	id, err := aio.submitIO(IOCmdPwrite, [][]byte{b}, offset)
	if err != nil {
		return 0, err
	}

	return aio.WaitFor(id)
}

func (aio *AsyncIO) Read(b []byte) (int, error) {
	nr, err := aio.Read(b)
	aio.offset += nr
	return nr, err
}

func (aio *AsyncIO) ReadAt(b []byte, offset int64) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	id, err := aio.submitIO(IOCmdPread, [][]byte{b}, offset)
	if err != nil {
		return 0, err
	}

	return aio.WaitFor(id)
}

func (aio *AsyncIO) WriteAtv(bs [][]byte, off int64) (int, error) {
	if bs == nil {
		return 0, nil
	}
	
	id, err := aio.submitIO(IOCmdPwritev, bs, offset)
	if err != nil {
		return 0, err
	}

	return aio.WaitFor(id)
}

func (aio *AsyncIO) submitIO(cmd IocbCmd, bs [][]byte, offset int64) (RequestID, error) {
	write := false

	// get the next available iocb
	iocb := aio.getNextReady()
	switch cmd {
	case IOCmdPread:
		iocb.PrepPread(bs[0], offset)
	case IOCmdPwrite:
		write = true
		iocb.PrepPwrite(bs[0], offset)
	case IOCmdPreadv:
		iocb.PrepPreadv(bs, offset)
	case IOCmdPwritev:
		write = true
		iocb.PrepPwritev(bs, offset)
	}

	if _, err := aio.ioctx.Submit([]iocb{*iocb}); err != nil {
		aio.available.Set(iocb, iocb)
		return 0, err
	}

	aio.Lock()
	aio.reqID++
	id := aio.reqID
	if write {
		aio.reCalcEnd(offset + int64(len(b)))
	}
	aio.Unlock()

	// add the iocb to the running event pool
	aio.running.Set(iocb, &runningEvent{
		// this prevents the gc from collecting the buffer
		data:b,
		iocb:iocb,
		reqID:id,
	})

	// add the request to the request pool
	aio.request.Set(id, &requestState{
		iocb:iocb,
		done:false,
	})

	return id, nil
}

func (aio *AsyncIO) Append(bs [][]byte) (int, error) {
	if bs == nil {
		return 0, nil
	}

	id, err := aio.submitIO(IOCmdPwritev, bs, aio.end)
	if err != nil {
		return 0, err
	}

	return aio.WaitFor(id)
}

func (aio *AsyncIO) Seek(offset int64, whence int) (int64, error) {
	var nOffset int64
	switch whence {
	case 0:
		nOffset = offset
	case 1:
		nOffset = int64(aio.offset) + offset
	case 2:
		nOffset = int64(aio.end) + offset
	}

	aio.offset = nOffset
	return nOffset, nil
}

// Truncate will wait for all submitted jobs to finish 
// trunctate the file to the designated size.
func (aio *AsyncIO) Truncate(size int64) error {
	// do we really need to wait all running IO completed?
	// what will happen write file when truncate?
	aio.waitAll()
	return aio.fd.Truncate(size)
}

// Sync will wait for all submitted jobs to finish and then sync
// the file descriptor.  Because the Linux kernel does not actually
// support Sync via the AIO interface we just issue a plain old sync
// via userland. No async here. Sync don't ack outstanding requests
func (aio *AsyncIO) Sync() error {
	// do we really need to wait all running IO completed?
	// what will happen write file when sync?
	aio.waitAll()
	return aio.fd.Sync()
}

// FLock async IO not impl Flock
func (aio *AsyncIO) FLock() error {
	return nil
}

// FUnlock async IO not impl FUnlock
func (aio *AsyncIO) FUnlock() error {
	return nil
}

// Option return options
func (aio *AsyncIO) Option() Options {
	return aio.opt
}

func pointer2string(p unsafe.Pointer) string {
	return fmt.Sprintf("%p", p)
}

func int2string(num int64) string {
	return fmt.Sprintf("%d", num)
}

//translate an error code to error
func lookupErrNo(errno int) error {
	return fmt.Errorf("Error %d", errno)
}