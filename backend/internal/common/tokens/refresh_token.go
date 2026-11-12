package tokens

import (
	"crypto/rand"
	"encoding/hex"
)

const nBytes = 32

func GetRefreshToken() (string, error) {
	bytes := make([]byte, nBytes)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
