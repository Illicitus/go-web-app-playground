package utils

import (
	"crypto/rand"
	"encoding/base64"
)

const RememberTokenBytes = 32

func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return b, nil
}

func NBytes(base64String string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// RememberToken is a helper function designed to generate remember tokens of a predetermined byte size.
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}