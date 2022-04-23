package utils

import (
	"io"
	"math"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

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

func CutString(arr []byte, len_line int) [][]byte {
	num_new_strings := int(math.Ceil(float64(len(arr)) / float64(len_line)))
	if len_line > 0 && num_new_strings > 0 {
		remaining_arr := arr[:]
		ret := make([][]byte, num_new_strings)
		for i := 0; i < num_new_strings; i++ {
			if len_line < len(remaining_arr) {
				ret[i] = remaining_arr[:len_line]
				remaining_arr = remaining_arr[len_line:]
			} else {
				ret[i] = remaining_arr[:]
			}
		}
		return ret
	} else {
		return [][]byte{}
	}
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func Clone(s string) string {
	b := make([]byte, len(s))
	copy(b, s)
	return *(*string)(unsafe.Pointer(&b))
}

func RandSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
