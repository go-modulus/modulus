package utils

import (
	"crypto/hmac"
	cryptoRand "crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func SafeDeref[T any](p *T) T {
	if p == nil {
		var v T
		return v
	}
	return *p
}

func RandomNumberString(length int) string {
	const runes = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

func RandomString(length int) string {
	const runes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

func HashString(str string) string {
	calculator := sha1.New()
	_, _ = calculator.Write([]byte(str))
	hash := calculator.Sum(nil)
	eb := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	base64.StdEncoding.Encode(eb, hash)

	return string(eb)
}

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := cryptoRand.Read(b); err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}

func HmacHash(str string, key string) string {
	keyForSign := []byte(key)
	h := hmac.New(sha1.New, keyForSign)
	_, _ = h.Write([]byte(str))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func ToStrings[T fmt.Stringer](l []T) []string {
	res := make([]string, len(l))
	for i, t := range l {
		res[i] = t.String()
	}
	return res
}

func ToP[T any](val T) *T {
	return &val
}

func FromP[T any](p *T) T {
	if p == nil {
		var v T
		return v
	}
	return *p
}
