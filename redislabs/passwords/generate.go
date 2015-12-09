package passwords

import (
	"crypto/rand"
	"io"
)

var (
	// Characters contains characters used for generating passwords.
	Characters = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+,.?/:;{}[]`~")
)

// Generate returns a randomly generated password of the given strength (length).
// Password characters are taken from the passwords.Characters set.
func Generate(length int) (string, error) {
	pass := []byte{}

	min, max := minMax(Characters)
	for len(pass) < length {
		chunk := make([]byte, length)
		_, err := io.ReadFull(rand.Reader, chunk)
		if err != nil {
			return "", err
		}
		for _, c := range chunk {
			if c > max || c < min {
				continue
			}
			pass = append(pass, c)
		}
	}
	return string(pass), nil
}

func minMax(bytes []byte) (byte, byte) {
	if len(bytes) == 0 {
		panic("minMax has been called with an empty byte array")
	}
	var (
		min = bytes[0]
		max = bytes[0]
	)
	for _, b := range bytes {
		if b < min {
			min = b
		} else if b > max {
			max = b
		}
	}
	return min, max
}
