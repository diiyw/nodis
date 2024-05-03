package redis

import "strconv"

func FormatFloat64(s string, inf float64) (float64, error) {
	var f string
	if s[0] == '(' {
		f = s[1:]
	} else {
		f = s
	}
	return strconv.ParseFloat(f, 64)
}

func FormatInt64(s string, inf int64) (int64, error) {
	var i string
	if s[0] == '(' {
		i = s[1:]
	} else {
		i = s
	}
	return strconv.ParseInt(i, 10, 64)
}
