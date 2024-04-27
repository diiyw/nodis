package nodis

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/diiyw/nodis/ds"
	"github.com/diiyw/nodis/ds/str"
	"github.com/diiyw/nodis/fs"
)

var (
	driver = &fs.Memory{}
)

func TestStorePut(t *testing.T) {
	tempDir := "testdata"
	_ = driver.RemoveAll(tempDir)
	// Create a new Store instance
	store := newStore(tempDir, 1024, 1000, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	ds := str.NewString()
	ds.Set(value)
	e := newEntry(key, ds, time.Now().Unix())
	// Call the put method
	err := store.putEntry(newEntry(key, ds, time.Now().Unix()))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Retrieve the value from the index
	idx, _ := store.keys.Get(key)
	if idx == nil {
		t.Fatalf("Failed to retrieve index for key: %s", key)
	}

	// Verify the index values
	if idx.fileId != 0 {
		t.Errorf("Expected fileID to be 0, got %d", idx.fileId)
	}
	if idx.offset != 0 {
		t.Errorf("Expected offset to be 0, got %d", idx.offset)
	}
	v, _ := e.Marshal()
	if idx.size != uint32(len(v)) {
		t.Errorf("Expected size to be %d, got %d", len(v), idx.size)
	}
	if idx.expiration <= 0 {
		t.Errorf("Expected expiration to be greater than 0, got %d", idx.expiration)
	}

	// Read the value from the aof file
	aofFile, err := driver.OpenFile(filepath.Join(tempDir, "nodis.0.aof"), os.O_RDONLY)
	if err != nil {
		t.Fatalf("Failed to open aof file: %v", err)
	}
	defer aofFile.Close()

	aofData, err := aofFile.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read aof file: %v", err)
	}

	if !bytes.Equal(aofData, v) {
		t.Errorf("Expected aof data to be %v, got %v", v, aofData)
	}
}

func TestStoreGet(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024, 1000, driver)

	// Generate a random name and value for testing
	name := "testKey"
	value := []byte("testValue")
	ds := str.NewString()
	ds.Set(value)
	// Call the put method
	err := store.putEntry(newEntry(name, ds, time.Now().Unix()))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}
	key, ok := store.keys.Get(name)
	if !ok {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	// Call the get method
	data, err := store.getEntryRaw(key)
	if err != nil {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	if data == nil {
		t.Fatalf("Failed to retrieve index for key: %s", name)
	}
}

func TestStoreMultiPut(t *testing.T) {
	tempDir := "testdata"
	_ = driver.RemoveAll(tempDir)
	// Create a new Store instance
	store := newStore(tempDir, 1024, 1000, driver)

	var kv = map[string][]byte{
		"testKey1": []byte("testValue1"),
		"testKey2": []byte("testValue2"),
		"testKey3": []byte(""),
		"testKey4": []byte("testValue4"),
		"testKey5": []byte("testValue5"),
	}
	result := make(map[string][]byte)
	for key, value := range kv {
		ds := str.NewString()
		ds.Set(value)
		// Call the put method
		e := newEntry(key, ds, time.Now().Unix()+3600)
		data, _ := e.Marshal()
		result[key] = data
		err := store.putEntry(e)
		if err != nil {
			t.Fatalf("Failed to put key-value pair: %v", err)
		}
	}

	if store.keys.Len() != 5 {
		t.Fatalf("Expected index count to be 5, got %d", store.keys.Len())
	}
	for name, value := range kv {
		key, ok := store.keys.Get(name)
		if !ok {
			t.Fatalf("Failed to get value for key: %v", name)
		}
		v, err := store.getEntryRaw(key)
		if err != nil {
			t.Fatalf("Failed to get value for key: %v", err)
		}
		if v == nil {
			t.Fatalf("Failed to retrieve index for key: %s", name)
		}

		if !bytes.Equal(result[name], v) {
			t.Errorf("Expected value to be %v, got %v", value, v)
		}
	}
}

func TestStoreMultiFilePut(t *testing.T) {
	tempDir := "testdata"
	_ = driver.RemoveAll(tempDir)
	os.Mkdir(tempDir, 0755)
	// Create a new Store instance
	store := newStore(tempDir, 10, 1000, driver)
	result := make(map[string][]byte)
	var kv = []map[string][]byte{
		{
			"testKey1": []byte("testValue11"),
		},
		{
			"testKey2": []byte("testValue22"),
		},
		{
			"testKey3": []byte(""),
		},
		{
			"testKey4": []byte("testValue44"),
		},
		{
			"testKey5": []byte("testValue55"),
		},
	}

	for _, m := range kv {
		for key, value := range m {
			ds := str.NewString()
			ds.Set(value)
			// Call the put method
			e := newEntry(key, ds, time.Now().Unix()+3600)
			data, _ := e.Marshal()
			result[key] = data
			err := store.putEntry(e)
			if err != nil {
				t.Fatalf("Failed to put key-value pair: %v", err)
			}
		}
	}

	for _, m := range kv {
		for name, value := range m {
			key, ok := store.keys.Get(name)
			if !ok {
				t.Fatalf("Failed to get value for key: %v", name)
			}
			v, err := store.getEntryRaw(key)
			if err != nil {
				t.Fatalf("Failed to get value for key: %v , err %v ", name, err)
			}
			if v == nil {
				t.Fatalf("Failed to retrieve index for key: %s", name)
			}

			if !bytes.Equal(result[name], v) {
				t.Errorf("Expected value to be %v, got %v", value, v)
			}
		}
	}
}

func TestStoreRemove(t *testing.T) {
	tempDir := "testdata"
	_ = driver.RemoveAll(tempDir)
	// Create a new Store instance
	store := newStore(tempDir, 1024, 1000, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	ds := str.NewString()
	ds.Set(value)
	// Call the put method
	err := store.putEntry(newEntry(key, ds, time.Now().Unix()+3600))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Call the remove method
	store.delKey(key)

	// Retrieve the value from the index
	idx, _ := store.keys.Get(key)
	if idx != nil {
		t.Fatalf("Expected index to be nil, got %v", idx)
	}
}

func TestStorePutRaw(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024, 1000, driver)

	// Generate a random name and value for testing
	name := "testKey"
	value := []byte("testValue")
	d := str.NewString()
	d.Set(value)
	expiration := time.Now().Unix() + 3600
	var e = newEntry(name, d, expiration)
	data, err := e.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}
	k := newKey()
	k.expiration = expiration
	// Call the putRaw method
	err = store.putRaw(name, k, data)
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Retrieve the value from the index
	key, _ := store.keys.Get(name)
	if key == nil {
		t.Fatalf("Failed to retrieve index for key: %s", name)
	}
	nd, err := store.getEntryRaw(key)
	if err != nil {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	if !bytes.Equal(nd, data) {
		t.Errorf("Expected value to be %v, got %v", data, nd)
	}
}

func TestStore_parseDs(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"))
	data := n.GetEntry("test")
	k, d, err := n.store.parseDs(data)
	if err != nil {
		t.Errorf("parseDs() = %v, want %v", err, nil)
	}
	if k != "test" {
		t.Errorf("parseDs() = %v, want %v", k, "test")
	}
	if d.Type() != ds.String {
		t.Errorf("parseDs() = %v, want %v", d.Type(), ds.String)
	}
	_ = n.store.close()
}
