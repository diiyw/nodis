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
	store := newStore(tempDir, 1024, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	strVal := str.NewString()
	strVal.Set(value)
	e := newValueEntry(key, strVal, time.Now().Unix())
	// Call the put method
	err := store.saveValueEntry(newValueEntry(key, strVal, time.Now().Unix()))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Retrieve the value from the index
	m, _ := store.metadata.Get(key)
	if m == nil {
		t.Fatalf("Failed to retrieve index for key: %s", key)
	}

	// Verify the index values
	if m.key.fileId != 0 {
		t.Errorf("Expected fileID to be 0, got %d", m.key.fileId)
	}
	if m.key.offset != 0 {
		t.Errorf("Expected offset to be 0, got %d", m.key.offset)
	}
	v := e.encode()
	if m.key.size != uint32(len(v)) {
		t.Errorf("Expected size to be %d, got %d", len(v), m.key.size)
	}
	if m.expiration <= 0 {
		t.Errorf("Expected expiration to be greater than 0, got %d", m.expiration)
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

	// Generate a random name and value for testing
	name := "testKey"
	value := []byte("testValue")
	strVal := str.NewString()
	strVal.Set(value)
	// Call the put method
	err := store.saveValueEntry(newValueEntry(name, strVal, time.Now().Unix()))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}
	meta, ok := store.metadata.Get(name)
	if !ok {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	// Call the get method
	data, err := store.getValueEntryRaw(meta.key)
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
		strVal := str.NewString()
		strVal.Set(value)
		// Call the put method
		e := newValueEntry(key, strVal, time.Now().Unix()+3600)
		data := e.encode()
		result[key] = data
		err := store.saveValueEntry(e)
		if err != nil {
			t.Fatalf("Failed to put key-value pair: %v", err)
		}
	}

	if store.metadata.Len() != 5 {
		t.Fatalf("Expected index count to be 5, got %d", store.metadata.Len())
	}
	for name, value := range kv {
		meta, ok := store.metadata.Get(name)
		if !ok {
			t.Fatalf("Failed to get value for key: %v", name)
		}
		v, err := store.getValueEntryRaw(meta.key)
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
	_ = os.Mkdir(tempDir, 0755)
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
			v := str.NewString()
			v.Set(value)
			// Call the put method
			e := newValueEntry(key, v, time.Now().Unix()+3600)
			data := e.encode()
			result[key] = data
			err := store.saveValueEntry(e)
			if err != nil {
				t.Fatalf("Failed to put key-value pair: %v", err)
			}
		}
	}

	for _, m := range kv {
		for name, value := range m {
			meta, ok := store.metadata.Get(name)
			if !ok {
				t.Fatalf("Failed to get value for key: %v", name)
			}
			v, err := store.getValueEntryRaw(meta.key)
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
	store := newStore(tempDir, 1024, driver)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")
	strVal := str.NewString()
	strVal.Set(value)
	// Call the put method
	err := store.saveValueEntry(newValueEntry(key, strVal, time.Now().Unix()+3600))
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}
}

func TestStorePutRaw(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024, driver)

	// Generate a random name and value for testing
	name := "testKey"
	value := []byte("testValue")
	strVal := str.NewString()
	strVal.Set(value)
	expiration := time.Now().Unix() + 3600
	var e = newValueEntry(name, strVal, expiration)
	data := e.encode()
	m := newMetadata()
	m.expiration = expiration
	// Call the saveValueRaw method
	err := store.saveValueRaw(name, m, data)
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Retrieve the value from the index
	meta, _ := store.metadata.Get(name)
	if meta == nil {
		t.Fatalf("Failed to retrieve index for key: %s", name)
	}
	nd, err := store.getValueEntryRaw(meta.key)
	if err != nil {
		t.Fatalf("Failed to get value for key: %v", err)
	}
	if !bytes.Equal(nd, data) {
		t.Errorf("Expected value to be %v, got %v", data, nd)
	}
}

func TestStore_parseValue(t *testing.T) {
	_ = os.RemoveAll("testdata")
	opt := &Options{
		Path: "testdata",
	}
	n := Open(opt)
	n.Set("test", []byte("test"), false)
	data := n.GetEntry("test")
	value, err := n.store.parseValue(data)
	if err != nil {
		t.Errorf("parseValue() = %v, want %v", err, nil)
	}
	if value.Type() != ds.String {
		t.Errorf("parseValue() = %v, want %v", value.Type(), ds.String)
	}
	_ = n.store.close()
}
