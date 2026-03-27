package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a password against its bcrypt hash.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateRandomHex generates a random hex string of the given byte length.
func GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SHA256Hex computes the SHA256 hash of a string.
func SHA256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

// GenerateFlag generates a flag string.
func GenerateFlag(format, teamID, service string, round int) string {
	random, _ := GenerateRandomHex(16)
	return fmt.Sprintf(format, fmt.Sprintf("%s_%s_%d_%s", teamID, service, round, random))
}

// GenerateToken generates a random API token.
func GenerateToken() (string, error) {
	return GenerateRandomHex(32)
}
