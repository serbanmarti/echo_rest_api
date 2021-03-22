package security

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/argon2"

	"echo_rest_api/internal"
)

const (
	threads = uint8(2)   // The number of available CPU's
	memory  = uint32(32) // The memory consumption of the hashing process in MB
	_time   = uint32(4)  // The number of passes over the memory
	keyLen  = uint32(32) // The number of bytes in the resulted password hash
)

// HashPassword hashes password using the given salt
func HashPassword(password string, salt []byte) string {
	// Convert the password to bytes
	bytesPassword := []byte(password)

	// Create the hash of the password
	hash := argon2.IDKey(bytesPassword, salt, _time, memory, threads, keyLen)

	// Return the string hashed password
	return base64.StdEncoding.EncodeToString(hash)
}

// NewSalt is the method for random byte slice generation
func NewSalt() ([]byte, error) {
	s := make([]byte, keyLen)

	if _, err := rand.Read(s); err != nil {
		return nil, internal.NewError(internal.ErrBEHashSalt, nil, 2)
	}

	return s, nil
}
