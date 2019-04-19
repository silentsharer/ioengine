// +build windows

package ioengine

// WriteAtv simulate writeatv by calling writev serially and dose not change the file offset.
func (fi *FileIO) WriteAtv(bs [][]byte, off int64) (int, error) {
	return genericWriteAtv(fi, bs, off)
}

// Append write data to the end of file.
// we recommend that open file with O_APPEND
func (fi *FileIO) Append(bs [][]byte) (int, error) {
	return genericAppend(fi, bs)
}
