package utils

import "unsafe"

// Byte2String no copy convert byte to string
func Byte2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
