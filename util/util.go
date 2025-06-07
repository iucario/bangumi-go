package util

import (
	"log/slog"
	"os/user"
	"strconv"
)

// Return -1 if not found
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

// Returns {HOME}/.config/bangumi-go
func ConfigDir() string {
	usr, err := user.Current()
	if err != nil {
		slog.Error(err.Error())
	}
	configDir := usr.HomeDir + "/.config/bangumi-go"
	return configDir
}
