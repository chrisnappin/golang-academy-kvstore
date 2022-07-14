// Package hash provides password hashing using argon2.
package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memoryUsedKB    = 64 * 1024
	iterations      = 3
	parallelism     = 2
	saltLengthBytes = 16
	hashLengthBytes = 32
)

var (
	errInvalidHash         = errors.New("the encoded hash is not in the correct format")
	errIncompatibleVersion = errors.New("incompatible version of argon2")
)

// GenerateHash generates an argon2 hash of the specified password, formatted as a string
// along with the salt and hash algorithm parameters.
//
// Because a cryptographically strong salt is randomly generated every time,
// this does not produce repeatable results.
func GenerateHash(password string) (string, error) {
	// use cryptographically strong salt
	salt, err := generateRandomBytes(saltLengthBytes)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterations, memoryUsedKB, parallelism, hashLengthBytes)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// format into the standard encoded hash representation.
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memoryUsedKB, iterations, parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

// VerifyAgainstHash checks the specified password against the hash, by generating a hash of the
// specified password using the salt and hash algorithm parameters from the encoded hash.
// The comparison is made in constant time to help prevent timing attacks. Returns whether the
// specified password matches the hash or not.
func VerifyAgainstHash(password string, encodedHash string) (bool, error) {
	memory, iterations, parallelism, keyLength, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func decodeHash(encodedHash string) (uint32, uint32, uint8, uint32, []byte, []byte, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return 0, 0, 0, 0, nil, nil, errInvalidHash
	}

	version := 0

	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return 0, 0, 0, 0, nil, nil, err
	}

	if version != argon2.Version {
		return 0, 0, 0, 0, nil, nil, errIncompatibleVersion
	}

	memory, iterations, parallelism := uint32(0), uint32(0), uint8(0)

	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return 0, 0, 0, 0, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return 0, 0, 0, 0, nil, nil, err
	}

	hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return 0, 0, 0, 0, nil, nil, err
	}

	keyLength := uint32(len(hash))

	return memory, iterations, parallelism, keyLength, salt, hash, nil
}

func generateRandomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}
