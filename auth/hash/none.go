package hash

type None struct{}

func NewNone() TokenHashStrategy {
	return &None{}
}

func (s *None) Hash(token string) string {
	return token
}
