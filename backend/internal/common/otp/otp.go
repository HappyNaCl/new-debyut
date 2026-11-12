package otp

import (
	"crypto/rand"
	"math/big"
)

func Generate() string {
	const digits = 6
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "000000"
	}

	out := make([]byte, digits)
	val := n.Int64()
	for i := digits - 1; i >= 0; i-- {
		out[i] = byte('0' + val%10)
		val /= 10
	}
	return string(out)
}
