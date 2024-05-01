package redis

import "strconv"

func Float64(s string, def ...float64) (float64, error) {
	if s == "" || s == "-inf" {
		return 0, nil
	}
	if s == "+inf" {
		return def[0], nil
	}
	if s[0] == '(' {
		v, err := strconv.ParseFloat(s[1:], 64)
		if err != nil {
			return 0, err
		}
		return v + 1, nil
	}
	return strconv.ParseFloat(s, 64)
}

func Int64(s string) (int64, error) {
	if s == "" || s == "-inf" {
		return 0, nil
	}
	if s == "+inf" {
		return -1, nil
	}
	if s[0] == '(' {
		v, err := strconv.ParseInt(s[1:], 10, 64)
		if err != nil {
			return 0, err
		}
		return v + 1, nil
	}
	return strconv.ParseInt(s, 10, 64)
}
