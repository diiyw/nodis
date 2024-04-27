package utils

import "unsafe"

func ToUpper(v string) string {
	buf := make([]byte, len(v))
	for i, vv := range v {
		if 'a' <= vv && vv <= 'z' {
			buf[i] = uint8(vv - 32)
		} else {
			buf[i] = uint8(vv)
		}
	}
	return unsafe.String(unsafe.SliceData(buf), len(buf))
}

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
