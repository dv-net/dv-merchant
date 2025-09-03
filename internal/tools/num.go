package tools

import (
	"crypto/rand"
	"math/big"
)

func RandomNumber(low, hi int) int {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-low)))
	return low + int(nBig.Int64())
}

func RandomCodeGenerator() int {
	return RandomNumber(100000, 999999)
}

func RandomSliceElement[T any](sl []T) T {
	return sl[RandomNumber(0, len(sl))]
}
