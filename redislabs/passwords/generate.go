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
	pass := make([]byte, length)

	for len(pass) < length {
		chunk := make([]byte, length)
		_, err := io.ReadFull(rand.Reader, chunk)
		if err != nil {
			return "", err
		}
		for _, c := range chunk {
			if c >= byte(len(Characters)) {
				continue
			}
			pass = append(pass, Characters[c%byte(len(Characters))])
		}
	}
	return string(pass), nil
}
