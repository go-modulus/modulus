package hash

import (
	"crypto/sha1"
	"encoding/base64"
)

type SHA1 struct{}

func NewSha1() TokenHashStrategy {
	return &SHA1{}
}

func (s *SHA1) Hash(token string) string {
	sha := sha1.New()
	sha.Write([]byte(token))
	return base64.StdEncoding.EncodeToString(sha.Sum(nil))
}
