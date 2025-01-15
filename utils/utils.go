package utils

import (
	"crypto/rand"
	"math/big"
)

var (
	LETTERS                   = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func RandStr(n int) string {
	b := make([]rune, n)
	max := big.NewInt(int64(len(LETTERS)) - 1)
	for i := range b {
		n, _ := rand.Int(rand.Reader, max)
		b[i] = LETTERS[n.Int64()]
	}
	return string(b)
}
