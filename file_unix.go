// +build linux darwin

package ioengine

// ReadAtv like linux preadv, read from the specifies offset and dose not change the file offset.
func (fi *FileIO) ReadAtv(off int64, bs ...[]byte) (int, error) {
	return 0, nil
}

// WriteAtv like linux pwritev, write to the specifies offset and dose not change the file offset.
func (fi *FileIO) WriteAtv(off int64, bs ...[]byte) (int, error) {
	return 0, nil
}

// Append write data to the end of file.
func (fi *FileIO) Append(bs ...[]byte) (int, error) {
	return 0, nil
}
