package hash

import (
	"bytes"
	"strconv"
	"testing"
)

func TestHash_HSet(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	v := hash.HGet(key)
	if v == nil {
		t.Errorf("HSet failed, expected %s but got nothing", value)
	}
	if !bytes.Equal(v, value) {
		t.Errorf("HSet failed, expected %s but got %s", value, v)
	}
}

func TestHash_HDel(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	hash.HDel(key)
	v := hash.HGet(key)
	if v != nil {
		t.Errorf("HDel failed, expected nothing but got %s", value)
	}
}

func TestHash_HLen(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	length := hash.HLen()
	if length != 1 {
		t.Errorf("HLen failed, expected 1 but got %d", length)
	}
}

func TestHash_HKeys(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	keys := hash.HKeys()
	if len(keys) != 1 {
		t.Errorf("HKeys failed, expected 1 but got %d", len(keys))
	}
	if keys[0] != key {
		t.Errorf("HKeys failed, expected %s but got %s", key, keys[0])
	}
}

func TestHash_HExists(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	ok := hash.HExists(key)
	if !ok {
		t.Errorf("HExists failed, expected true but got false")
	}
}

func TestHash_HGetAll(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	values := hash.HGetAll()
	if len(values) != 1 {
		t.Errorf("HGetAll failed, expected 1 but got %d", len(values))
	}
	if !bytes.Equal(values[key], value) {
		t.Errorf("HGetAll failed, expected %s but got %s", value, values[key])
	}
}

func TestHash_HIncrBy(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	var value int64 = 1

	hash.HIncrBy(key, value)
	v := hash.HGet(key)
	if v == nil {
		t.Errorf("HIncrBy failed, expected %d but got nothing", value)
	}
	if string(v) != strconv.FormatInt(value, 10) {
		t.Errorf("HIncrBy failed, expected %d but got %d", value, v)
	}
}

func TestHash_HIncrBy2(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	var value int64 = 1

	hash.HIncrBy(key, value)
	hash.HIncrBy(key, value)
	v := hash.HGet(key)
	if v == nil {
		t.Errorf("HIncrBy failed, expected %d but got nothing", value*2)
	}
	if string(v) != strconv.FormatInt(value*2, 10) {
		t.Errorf("HIncrBy failed, expected %d but got %d", value*2, v)
	}
}

func TestHash_HIncrByFloat(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := 1.0

	hash.HIncrByFloat(key, value)
	v := hash.HGet(key)
	if v == nil {
		t.Errorf("HIncrByFloat failed, expected %f but got nothing", value)
	}
	if string(v) != strconv.FormatFloat(value, 'f', -1, 64) {
		t.Errorf("HIncrByFloat failed, expected %f but got %s", value, string(v))
	}
}

func TestHash_HIncrByFloat2(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := 1.0

	hash.HIncrByFloat(key, value)
	hash.HIncrByFloat(key, value)
	v := hash.HGet(key)
	if v == nil {
		t.Errorf("HIncrByFloat failed, expected %f but got nothing", value*2)
	}
	if string(v) != strconv.FormatFloat(value*2, 'f', -1, 64) {
		t.Errorf("HIncrByFloat failed, expected %f but got %s", value*2, string(v))
	}
}

func TestHash_HMSet(t *testing.T) {
	hash := NewHashMap()
	values := map[string][]byte{
		"testKey1": []byte("testValue1"),
		"testKey2": []byte("testValue2"),
	}

	hash.HMSet(values)
	v := hash.HGet("testKey1")
	if v == nil {
		t.Errorf("HMSet failed, expected %s but got nothing", "testValue1")
	}
	if string(v) != "testValue1" {
		t.Errorf("HMSet failed, expected %s but got %s", "testValue1", v)
	}
	v = hash.HGet("testKey2")
	if v == nil {
		t.Errorf("HMSet failed, expected %s but got nothing", "testValue2")
	}
	if string(v) != "testValue2" {
		t.Errorf("HMSet failed, expected %s but got %s", "testValue2", v)
	}
}

func TestHash_HSetNX(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	ok := hash.HSetNX(key, value)
	if !ok {
		t.Errorf("HSetNX failed, expected true but got false")
	}
	ok = hash.HSetNX(key, value)
	if ok {
		t.Errorf("HSetNX failed, expected false but got true")
	}
	v := hash.HGet("testKey")
	if v == nil {
		t.Errorf("HMSet failed, expected %s but got nothing", "testValue2")
	}
	if string(v) != "testValue" {
		t.Errorf("HMSet failed, expected %s but got %s", "testValue", v)
	}
}

func TestHash_HVals(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	values := hash.HVals()
	if len(values) != 1 {
		t.Errorf("HVals failed, expected 1 but got %d", len(values))
	}
	if string(values[0]) != string(value) {
		t.Errorf("HVals failed, expected %s but got %s", value, values[0])
	}
}

func TestHash_HScan(t *testing.T) {
	hash := NewHashMap()
	key := "testKey"
	value := []byte("testValue")

	hash.HSet(key, value)
	_, values := hash.HScan(0, "*", 1)
	if len(values) != 1 {
		t.Errorf("HScan failed, expected 1 but got %d", len(values))
	}
	if !bytes.Equal(values[key], value) {
		t.Errorf("HScan failed, expected %s but got %s", value, values[key])
	}
}

func TestHash_HScan2(t *testing.T) {
	hash := NewHashMap()
	values := map[string][]byte{
		"testKey1": []byte("testValue1"),
		"testKey2": []byte("testValue2"),
		"testKey3": []byte("testValue3"),
	}

	hash.HMSet(values)
	if hash.HLen() != 3 {
		t.Errorf("HScan failed, expected 3 but got %d", hash.HLen())
	}
	cursor, v := hash.HScan(0, "*", 2)
	if len(v) != 2 {
		t.Errorf("HScan failed, expected 2 but got %d", len(v))
	}
	_, v = hash.HScan(cursor, "*", 1)
	if len(v) != 1 {
		t.Errorf("HScan failed, expected 1 but got %d", len(v))
	}
}

func BenchmarkHashMap_HMGet(b *testing.B) {
	hash := NewHashMap()
	values := map[string][]byte{
		"testKey1": []byte("testValue1"),
		"testKey2": []byte("testValue2"),
		"testKey3": []byte("testValue3"),
	}

	hash.HMSet(values)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.HMGet("testKey1", "testKey2", "testKey3")
	}
}

func TestGetSetValue(t *testing.T) {
	h := NewHashMap()
	h.HSet("testKey", []byte("testValue"))
	h.HSet("testKey1", []byte("testValue1"))
	h.HSet("testKey2", []byte("testValue2"))
	v := h.GetValue()
	h2 := NewHashMap()
	h2.SetValue(v)
	vv := h.HGet("testKey")
	if vv == nil {
		t.Errorf("GetSetValue failed, expected %s but got nothing", "testValue")
	}
	if string(vv) != "testValue" {
		t.Errorf("GetSetValue failed, expected %s but got %s", "testValue", vv)
	}
	vv = h.HGet("testKey1")
	if vv == nil {
		t.Errorf("GetSetValue failed, expected %s but got nothing", "testValue")
	}
	if string(vv) != "testValue1" {
		t.Errorf("GetSetValue failed, expected %s but got %s", "testValue1", vv)
	}
	vv = h.HGet("testKey2")
	if vv == nil {
		t.Errorf("GetSetValue failed, expected %s but got nothing", "testValue")
	}
	if string(vv) != "testValue2" {
		t.Errorf("GetSetValue failed, expected %s but got %s", "testValue2", vv)
	}
}
