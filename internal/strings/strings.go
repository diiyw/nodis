package strings

import (
	"strings"
	"unsafe"
)

func ToUpper(v string) string { return strings.ToUpper(v) }

func Fnv32(key string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		h *= 16777619
		h ^= uint32(key[i])
	}
	return h
}

func String2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
