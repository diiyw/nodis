package storage

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/diiyw/nodis/ds"
	"github.com/tidwall/btree"
)

type Key struct {
	offset int64
	size   uint32
	fileId uint16
}

type Disk struct {
	fileSize  int64
	fileId    uint16
	path      string
	current   string
	indexFile string
	aof       *os.File
	keys      btree.Map[string, *Key]
}

func NewDisk(path string, fileSize int64) *Disk {
	return &Disk{
		path:      path,
		fileSize:  fileSize,
		indexFile: filepath.Join(path, "nodid.index"),
	}
}

// Open initializes the storage.
func (d *Disk) Open() error {
	err := os.MkdirAll(d.path, 0755)
	if err != nil {
		return err
	}
	return nil
}

// Get returns a value from the storage.
func (d *Disk) Get(key string) (ds.Value, error) {
	k, ok := d.keys.Get(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	if k.fileId == d.fileId {
		data := make([]byte, k.size)
		_, err := d.aof.ReadAt(data, k.offset)
		if err != nil {
			return nil, err
		}
		return parseValue(data)
	}
	// read from other file
	file := filepath.Join(d.path, "nodid."+strconv.Itoa(int(k.fileId))+".aof")
	f, err := os.OpenFile(file, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Println("Close file error: ", err)
		}
	}()
	data := make([]byte, k.size)
	_, err = f.ReadAt(data, k.offset)
	if err != nil {
		return nil, err
	}
	return parseValue(data)
}

// Put sets a value in the storage.
func (d *Disk) Put(key string, value ds.Value, expiration int64) error {
	data := NewValueEntry(key, value, expiration).encode()
	offset, err := d.check()
	if err != nil {
		return err
	}
	k := &Key{
		fileId: d.fileId,
		offset: offset,
		size:   uint32(len(data)),
	}
	d.keys.Set(key, k)
	_, err = d.aof.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a value from the storage.
func (d *Disk) Delete(key string) error {
	d.keys.Delete(key)
	return nil
}

func (d *Disk) GetIndex() []byte {
	fi, err := os.OpenFile(d.indexFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	data, err := d.readAll(fi)
	if err != nil {
		panic(err)
	}
	if len(data) > 2 {
		d.fileId = binary.LittleEndian.Uint16(data[:2])
		data = data[2:]
	}
	d.current = filepath.Join(d.path, "nodid."+strconv.Itoa(int(d.fileId))+".aof")
	d.aof, err = os.OpenFile(d.current, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		panic(err)
	}
	return data
}

// PutIndex sets the index.
func (d *Disk) PutIndex(index []byte) error {
	idxFile, err := os.OpenFile(d.indexFile+"~", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	var header = make([]byte, 2)
	binary.LittleEndian.PutUint16(header, d.fileId)
	_, err = idxFile.Write(header)
	if err != nil {
		return err
	}
	defer func() {
		if err = idxFile.Close(); err != nil {
			log.Println("Close sync error: ", err)
		}
	}()
	n, err := idxFile.Write(index)
	if err != nil {
		return err
	}
	if n != len(index) {
		return io.ErrShortWrite
	}
	err = os.Rename(d.indexFile+"~", d.indexFile)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the storage.
func (d *Disk) Close() error {
	err := d.aof.Close()
	if err != nil {
		return err
	}
	return nil
}

func (d *Disk) readAll(fi *os.File) ([]byte, error) {
	l, err := d.getFileSize(fi)
	if err != nil {
		return nil, err
	}
	size := int(l)
	if size == 0 {
		return nil, nil
	}
	data := make([]byte, size)
	_, err = d.readAt(fi, data, 0)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *Disk) readAt(fi *os.File, b []byte, off int64) (n int, err error) {
	n, err = fi.ReadAt(b, off)
	if err == io.EOF {
		err = nil
	}
	return n, err
}

func (d *Disk) getFileSize(fi *os.File) (int64, error) {
	fs, err := fi.Stat()
	if err != nil {
		return 0, err
	}
	return fs.Size(), nil
}

func (d *Disk) check() (int64, error) {
	var offset, err = d.getFileSize(d.aof)
	if err != nil {
		return 0, err
	}
	if offset >= d.fileSize {
		err = d.aof.Close()
		if err != nil {
			return 0, err
		}
		// open file with new file id
		d.fileId++
		d.current = filepath.Join(d.path, "nodid."+strconv.Itoa(int(d.fileId))+".aof")
		d.aof, err = os.OpenFile(d.current, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return 0, err
		}
		// update index file
		idxFi, err := os.OpenFile(d.indexFile, os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			return 0, err
		}
		defer func() {
			err := idxFi.Close()
			if err != nil {
				log.Println("Close index file error: ", err)
			}
		}()
		var header = make([]byte, 4)
		binary.LittleEndian.PutUint16(header, d.fileId)
		_, err = idxFi.WriteAt(header, 0)
		if err != nil {
			return 0, err
		}
		offset = 0
	}
	return offset, nil
}

func (d *Disk) Reset() error {
	err := d.aof.Close()
	if err != nil {
		return err
	}
	d.keys.Clear()
	err = os.Truncate(d.aof.Name(), 0)
	if err != nil {
		return err
	}
	d.aof, err = os.OpenFile(d.aof.Name(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	return err
}
