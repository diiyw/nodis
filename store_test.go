package nodis

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	store := newStore(tempDir, 1024, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	ds := str.NewString()
	ds.Set(value)
	e := newEntity(key, ds, time.Now().Unix())
	// Call the put method
	err := store.put(newEntity(key, ds, time.Now().Unix()))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Retrieve the value from the index
	idx, _ := store.index.Get(key)
	if idx == nil {
		t.Fatalf("Failed to retrieve index for key: %s", key)
	}

	// Verify the index values
	if idx.fileID != 0 {
		t.Errorf("Expected fileID to be 0, got %d", idx.fileID)
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
	store := newStore(tempDir, 1024, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	ds := str.NewString()
	ds.Set(value)
	// Call the put method
	err := store.put(newEntity(key, ds, time.Now().Unix()))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}
	// Call the get method
	data, err := store.get(key)
	if err != nil {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	if data == nil {
		t.Fatalf("Failed to retrieve index for key: %s", key)
	}
	_, d, _, err := parseDs(data)
	if err != nil || d == nil {
		t.Fatalf("Failed to parse value for key: %v", err)
	}
	v := d.(*str.String).Get()
	if !bytes.Equal(value, v) {
		t.Errorf("Expected value to be %v, got %v", value, v)
	}
}

func TestStoreMultiPut(t *testing.T) {
	tempDir := "testdata"
	_ = driver.RemoveAll(tempDir)
	// Create a new Store instance
	store := newStore(tempDir, 1024, driver)

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
		e := newEntity(key, ds, time.Now().Unix()+3600)
		data, _ := e.Marshal()
		result[key] = data
		err := store.put(e)
		if err != nil {
			t.Fatalf("Failed to put key-value pair: %v", err)
		}
	}

	if store.index.Len() != 5 {
		t.Fatalf("Expected index count to be 5, got %d", store.index.Len())
	}
	for key, value := range kv {
		v, err := store.get(key)
		if err != nil {
			t.Fatalf("Failed to get value for key: %v", err)
		}
		if v == nil {
			t.Fatalf("Failed to retrieve index for key: %s", key)
		}

		if !bytes.Equal(result[key], v) {
			t.Errorf("Expected value to be %v, got %v", value, v)
		}
	}
}

func TestStoreMultiFilePut(t *testing.T) {
	tempDir := "testdata"
	_ = driver.RemoveAll(tempDir)
	os.Mkdir(tempDir, 0755)
	// Create a new Store instance
	store := newStore(tempDir, 10, driver)
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
			e := newEntity(key, ds, time.Now().Unix()+3600)
			data, _ := e.Marshal()
			result[key] = data
			err := store.put(e)
			if err != nil {
				t.Fatalf("Failed to put key-value pair: %v", err)
			}
		}
	}

	for _, m := range kv {
		for key, value := range m {
			v, err := store.get(key)
			if err != nil {
				t.Fatalf("Failed to get value for key: %v , err %v ", key, err)
			}
			if v == nil {
				t.Fatalf("Failed to retrieve index for key: %s", key)
			}

			if !bytes.Equal(result[key], v) {
				t.Errorf("Expected value to be %v, got %v", value, v)
			}
		}
	}
}

func TestStoreRemove(t *testing.T) {
	tempDir := "testdata"
	_ = driver.RemoveAll(tempDir)
	// Create a new Store instance
	store := newStore(tempDir, 1024, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	ds := str.NewString()
	ds.Set(value)
	// Call the put method
	err := store.put(newEntity(key, ds, time.Now().Unix()+3600))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Call the remove method
	store.remove(key)

	// Retrieve the value from the index
	idx, _ := store.index.Get(key)
	if idx != nil {
		t.Fatalf("Expected index to be nil, got %v", idx)
	}
}

func TestStorePutRaw(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	d := str.NewString()
	d.Set(value)
	var e = newEntity(key, d, time.Now().Unix()+3600)
	data, err := e.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}
	// Call the putRaw method
	err = store.putRaw(key, data, e.Expiration)
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Retrieve the value from the index
	idx, _ := store.index.Get(key)
	if idx == nil {
		t.Fatalf("Failed to retrieve index for key: %s", key)
	}
	nd, err := store.get(key)
	if err != nil {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	if !bytes.Equal(nd, data) {
		t.Errorf("Expected value to be %v, got %v", data, nd)
	}
	_, dv, _, err := parseDs(data)
	if err != nil {
		t.Fatalf("Failed to parse value for key: %v", err)
	}
	v := dv.(*str.String).Get()
	if !bytes.Equal(value, v) {
		t.Errorf("Expected value to be %v, got %v", value, v)
	}
}
