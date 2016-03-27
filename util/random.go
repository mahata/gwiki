package util

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

func GetRandomString() string {
	b := make([]byte, 32)
	rand.Read(b)

	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
}
