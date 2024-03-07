package strparse

import (
	"strconv"
	"strings"
)

func StringToUint64Slice(str string) []uint64 {
	var slice []uint64
	strSlice := strings.Split(str, ",")

	for _, s := range strSlice {
		num, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			panic(err)
		}
		slice = append(slice, num)
	}

	return slice
}
