package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
)

// RandomHex return random hex string.
func RandomHex(size int) (string, error) {
	randByte := make([]byte, size)
	if _, err := rand.Read(randByte); err != nil {
		return "", err
	}
	return hex.EncodeToString(randByte), nil
}

// RandomUint64 return random uint64.
func RandomUint64() (uint64, error) {
	randByte := make([]byte, 8)
	if _, err := rand.Read(randByte); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(randByte), nil
}
