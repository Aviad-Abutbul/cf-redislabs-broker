package passwords

import (
	"crypto/rand"
	"math/big"
)

var (
	// Characters contains characters used for generating passwords.
	Characters = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+,.?/:;{}[]`~")
)

// Generate returns a randomly generated password of the given strength (length).
// Password characters are taken from the passwords.Characters set.
func Generate(length int) (string, error) {
	pass := []byte{}
	for len(pass) < length {
		pos, err := rand.Int(rand.Reader, big.NewInt(int64(len(Characters))))
		if err != nil {
			return "", err
		}
		pass = append(pass, Characters[pos.Int64()])
	}
	return string(pass), nil
}
