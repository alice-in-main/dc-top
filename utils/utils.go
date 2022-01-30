package utils

func FindByte(needle byte, haystack []byte) int {
	for index, item := range haystack {
		if item == needle {
			return index
		}
	}
	return -1
}
