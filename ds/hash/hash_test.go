package hash

import "testing"

func TestHash_HSet(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	hash.HSet(key, value)
	v, ok := hash.HGet(key)
	if !ok {
		t.Errorf("HSet failed, expected %s but got nothing", value)
	}
	if v != value {
		t.Errorf("HSet failed, expected %s but got %s", value, v)
	}
}

func TestHash_HDel(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	hash.HSet(key, value)
	hash.HDel(key)
	_, ok := hash.HGet(key)
	if ok {
		t.Errorf("HDel failed, expected nothing but got %s", value)
	}
}

func TestHash_HLen(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	hash.HSet(key, value)
	length := hash.HLen()
	if length != 1 {
		t.Errorf("HLen failed, expected 1 but got %d", length)
	}
}

func TestHash_HKeys(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

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
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	hash.HSet(key, value)
	ok := hash.HExists(key)
	if !ok {
		t.Errorf("HExists failed, expected true but got false")
	}
}

func TestHash_HGetAll(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	hash.HSet(key, value)
	values := hash.HGetAll()
	if len(values) != 1 {
		t.Errorf("HGetAll failed, expected 1 but got %d", len(values))
	}
	if values[key] != value {
		t.Errorf("HGetAll failed, expected %s but got %s", value, values[key])
	}
}

func TestHash_HIncrBy(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := 1

	hash.HIncrBy(key, value)
	v, ok := hash.HGet(key)
	if !ok {
		t.Errorf("HIncrBy failed, expected %d but got nothing", value)
	}
	if v != value {
		t.Errorf("HIncrBy failed, expected %d but got %d", value, v)
	}
}

func TestHash_HIncrBy2(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := 1

	hash.HIncrBy(key, value)
	hash.HIncrBy(key, value)
	v, ok := hash.HGet(key)
	if !ok {
		t.Errorf("HIncrBy failed, expected %d but got nothing", value*2)
	}
	if v != value*2 {
		t.Errorf("HIncrBy failed, expected %d but got %d", value*2, v)
	}
}

func TestHash_HIncrByFloat(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := 1.0

	hash.HIncrByFloat(key, value)
	v, ok := hash.HGet(key)
	if !ok {
		t.Errorf("HIncrByFloat failed, expected %f but got nothing", value)
	}
	if v != value {
		t.Errorf("HIncrByFloat failed, expected %f but got %f", value, v)
	}
}

func TestHash_HIncrByFloat2(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := 1.0

	hash.HIncrByFloat(key, value)
	hash.HIncrByFloat(key, value)
	v, ok := hash.HGet(key)
	if !ok {
		t.Errorf("HIncrByFloat failed, expected %f but got nothing", value*2)
	}
	if v != value*2 {
		t.Errorf("HIncrByFloat failed, expected %f but got %f", value*2, v)
	}
}

func TestHash_HMSet(t *testing.T) {
	hash := NewHash()
	values := map[string]any{
		"testKey1": "testValue1",
		"testKey2": "testValue2",
	}

	hash.HMSet(values)
	v, ok := hash.HGet("testKey1")
	if !ok {
		t.Errorf("HMSet failed, expected %s but got nothing", "testValue1")
	}
	if v != "testValue1" {
		t.Errorf("HMSet failed, expected %s but got %s", "testValue1", v)
	}
	v, ok = hash.HGet("testKey2")
	if !ok {
		t.Errorf("HMSet failed, expected %s but got nothing", "testValue2")
	}
	if v != "testValue2" {
		t.Errorf("HMSet failed, expected %s but got %s", "testValue2", v)
	}
}

func TestHash_HSetNX(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	ok := hash.HSetNX(key, value)
	if !ok {
		t.Errorf("HSetNX failed, expected true but got false")
	}
	ok = hash.HSetNX(key, value)
	if ok {
		t.Errorf("HSetNX failed, expected false but got true")
	}
	v, ok := hash.HGet("testKey")
	if !ok {
		t.Errorf("HMSet failed, expected %s but got nothing", "testValue2")
	}
	if v != "testValue" {
		t.Errorf("HMSet failed, expected %s but got %s", "testValue", v)
	}
}

func TestHash_HVals(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	hash.HSet(key, value)
	values := hash.HVals()
	if len(values) != 1 {
		t.Errorf("HVals failed, expected 1 but got %d", len(values))
	}
	if values[0] != value {
		t.Errorf("HVals failed, expected %s but got %s", value, values[0])
	}
}

func TestHash_HScan(t *testing.T) {
	hash := NewHash()
	key := "testKey"
	value := "testValue"

	hash.HSet(key, value)
	_, values := hash.HScan(0, "*", 1)
	if len(values) != 1 {
		t.Errorf("HScan failed, expected 1 but got %d", len(values))
	}
	if values[key] != value {
		t.Errorf("HScan failed, expected %s but got %s", value, values[key])
	}
}

func TestHash_HScan2(t *testing.T) {
	hash := NewHash()
	values := map[string]any{
		"testKey1": "testValue1",
		"testKey2": "testValue2",
		"testKey3": "testValue3",
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
