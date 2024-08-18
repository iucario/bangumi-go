package util

import "strconv"

func IndexOfString(arr []string, s string) int {
	for i, v := range arr {
		if v == s {
			return i
		}
	}
	return -1
}

func Uint32ToString(i uint32) string {
	return strconv.FormatUint(uint64(i), 10)
}
