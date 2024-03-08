package nodis

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStorePut(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")

	// Call the put method
	err := store.put(key, value, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Retrieve the value from the index
	index, _ := store.index.Get(key)
	if index == nil {
		t.Fatalf("Failed to retrieve index for key: %s", key)
	}

	// Verify the index values
	if index.ChunkID != 0 {
		t.Errorf("Expected ChunkID to be 0, got %d", index.ChunkID)
	}
	if index.Offset != 0 {
		t.Errorf("Expected Offset to be 0, got %d", index.Offset)
	}
	if index.Size != uint32(len(value)) {
		t.Errorf("Expected Size to be %d, got %d", len(value), index.Size)
	}
	if index.ExpiredAt <= 0 {
		t.Errorf("Expected ExpiredAt to be greater than 0, got %d", index.ExpiredAt)
	}

	// Read the value from the aof file
	aofFile, err := os.Open(filepath.Join(tempDir, "nodis.0.aof"))
	if err != nil {
		t.Fatalf("Failed to open aof file: %v", err)
	}
	defer aofFile.Close()

	aofData, err := io.ReadAll(aofFile)
	if err != nil {
		t.Fatalf("Failed to read aof file: %v", err)
	}

	if !bytes.Equal(aofData, value) {
		t.Errorf("Expected aof data to be %v, got %v", value, aofData)
	}
}

func TestStoreGet(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")

	// Write the value to the aof file
	aofFile, err := os.Create(filepath.Join(tempDir, "nodis.0.aof"))
	if err != nil {
		t.Fatalf("Failed to create aof file: %v", err)
	}
	defer aofFile.Close()

	_, err = aofFile.Write(value)
	if err != nil {
		t.Fatalf("Failed to write value to aof file: %v", err)
	}

	// Create an index entry for the key
	store.index.Put(key, &index{
		ChunkID:   0,
		Offset:    0,
		Size:      uint32(len(value)),
		ExpiredAt: time.Now().Unix(),
	})

	// Call the get method
	data, err := store.get(key)
	if err != nil {
		t.Fatalf("Failed to get value for key: %v", err)
	}

	if !bytes.Equal(data, value) {
		t.Errorf("Expected data to be %v, got %v", value, data)
	}
}

func TestStoreMultiPut(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024)

	var kv = map[string][]byte{
		"testKey1": []byte("testValue1"),
		"testKey2": []byte("testValue2"),
		"testKey3": []byte(""),
		"testKey4": []byte("testValue4"),
		"testKey5": []byte("testValue5"),
	}

	for key, value := range kv {
		err := store.put(key, value, time.Now().Unix()+3600)
		if err != nil {
			t.Fatalf("Failed to put key-value pair: %v", err)
		}
	}

	if store.index.Count() != 5 {
		t.Fatalf("Expected index count to be 5, got %d", store.index.Count())
	}
	for key, value := range kv {
		v, err := store.get(key)
		if err != nil {
			t.Fatalf("Failed to get value for key: %v", err)
		}
		if v == nil {
			t.Fatalf("Failed to retrieve index for key: %s", key)
		}

		if !bytes.Equal(value, v) {
			t.Errorf("Expected value to be %v, got %v", value, v)
		}
	}
}

func TestStoreMultiFilePut(t *testing.T) {
	tempDir := "testdata"
	os.RemoveAll(tempDir)
	os.Mkdir(tempDir, 0755)
	// Create a new Store instance
	store := newStore(tempDir, 10)

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
			err := store.put(key, value, time.Now().Unix()+3600)
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

			if !bytes.Equal(value, v) {
				t.Errorf("Expected value to be %v, got %v", value, v)
			}
		}
	}
}

func TestStoreRemove(t *testing.T) {
	tempDir := "testdata"

	// Create a new Store instance
	store := newStore(tempDir, 1024)

	// Generate a random key and value for testing
	key := "testKey"
	value := []byte("testValue")

	// Call the put method
	err := store.put(key, value, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to put key-value pair: %v", err)
	}

	// Call the remove method
	store.remove(key)

	// Retrieve the value from the index
	index, _ := store.index.Get(key)
	if index != nil {
		t.Fatalf("Expected index to be nil, got %v", index)
	}
}
