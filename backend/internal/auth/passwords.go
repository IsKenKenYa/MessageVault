package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func hashPassword(password string) (salt string, hash string, err error) {
	rawSalt := make([]byte, 16)
	if _, err := rand.Read(rawSalt); err != nil {
		return "", "", err
	}
	salt = hex.EncodeToString(rawSalt)
	sum := derivePasswordPBKDF2(password, salt, 120000)
	return salt, "pbkdf2$" + hex.EncodeToString(sum[:]), nil
}

func verifyPassword(password, salt, expected string) bool {
	var sum [32]byte
	if strings.HasPrefix(expected, "pbkdf2$") {
		sum = derivePasswordPBKDF2(password, salt, 120000)
		return hmac.Equal([]byte("pbkdf2$"+hex.EncodeToString(sum[:])), []byte(expected))
	}
	stripped := strings.TrimPrefix(expected, "sha256$")
	sum = derivePasswordLegacy(password, salt)
	return hmac.Equal([]byte(hex.EncodeToString(sum[:])), []byte(stripped))
}

func NeedsRehash(hash string) bool {
	return !strings.HasPrefix(hash, "pbkdf2$")
}

func derivePasswordPBKDF2(password, salt string, iterations int) [32]byte {
	prf := hmac.New(sha256.New, []byte(password))
	var result [32]byte
	prf.Write([]byte(salt))
	prf.Write([]byte{0, 0, 0, 1})
	u := prf.Sum(nil)
	copy(result[:], u)
	for i := 1; i < iterations; i++ {
		prf.Reset()
		prf.Write(u)
		u = prf.Sum(nil)
		for j := 0; j < 32; j++ {
			result[j] ^= u[j]
		}
	}
	return result
}

func derivePasswordLegacy(password, salt string) [32]byte {
	input := []byte(password + ":" + salt)
	sum := sha256.Sum256(input)
	for i := 0; i < 120000; i++ {
		sum = sha256.Sum256(append(sum[:], input...))
	}
	return sum
}
