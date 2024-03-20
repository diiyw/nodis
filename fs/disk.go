package fs

import (
	"io"
	"os"
)

type Disk struct {
}

type DiskFile struct {
	*os.File
}

func (d *Disk) OpenFile(filename string, flag int) (File, error) {
	fi, err := os.OpenFile(filename, flag, 0644)
	if err != nil {
		return nil, err
	}
	fi.Fd()
	return &DiskFile{File: fi}, nil
}

func (d *Disk) MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func (d *Disk) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (d *DiskFile) ReadAt(b []byte, off int64) (n int, err error) {
	n, err = d.File.ReadAt(b, off)
	if err == io.EOF {
		err = nil
	}
	return n, err
}

func (d *DiskFile) Write(b []byte) (n int, err error) {
	return d.File.Write(b)
}

func (d *DiskFile) Close() error {
	err := d.File.Sync()
	if err != nil {
		return err
	}
	return d.File.Close()
}

func (d *DiskFile) FileSize() (int64, error) {
	fi, err := d.File.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (d *DiskFile) Truncate(size int64) error {
	return d.File.Truncate(size)
}

func (d *DiskFile) ReadAll() ([]byte, error) {
	l, err := d.FileSize()
	if err != nil {
		return nil, err
	}
	size := int(l)
	if size == 0 {
		return nil, nil
	}
	data := make([]byte, size)
	_, err = d.ReadAt(data, 0)
	if err != nil {
		return nil, err
	}
	return data, nil
}
