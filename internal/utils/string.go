package utils

import (
	"strconv"
	"strings"
)

func SplitAndParseToInt(input string) (output []int) {
	tokens := strings.Split(input, "-")
	output = make([]int, len(tokens))
	for i, t := range tokens {
		val, err := strconv.ParseInt(t, 10, 32)
		if err != nil {
			return nil
		}
		output[i] = int(val)
	}

	return output
}
