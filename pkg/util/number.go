package util

import "strconv"

func BytesToInt(value []byte) int {
	heightInt, _ := strconv.Atoi(string(value))

	return heightInt
}
