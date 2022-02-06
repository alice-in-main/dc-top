package utils

func FindByte(needle byte, haystack []byte) int {
	for index, item := range haystack {
		if item == needle {
			return index
		}
	}
	return -1
}

func ConcatRuneArrays(slices [][]rune) []rune {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	tmp := make([]rune, totalLen)
	var i int
	for _, s := range slices {
		i += copy(tmp[i:], s)
	}
	return tmp
}
