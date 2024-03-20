package fs

type Fs interface {
	OpenFile(filename string, flag int) (File, error)
	MkdirAll(path string) error
	Rename(oldpath, newpath string) error
}

type File interface {
	ReadAt(b []byte, off int64) (n int, err error)
	Write(b []byte) (n int, err error)
	WriteAt(b []byte, off int64) (n int, err error)
	Close() error
	FileSize() (int64, error)
	Truncate(size int64) error
	ReadAll() ([]byte, error)
}
