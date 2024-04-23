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
