package util

import "os"

func IsNotExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
	}
	return false
}
