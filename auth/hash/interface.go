package hash

type TokenHashStrategy interface {
	Hash(token string) string
}
