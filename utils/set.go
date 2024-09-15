package utils

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](values ...T) Set[T] {
	set := Set[T]{}
	set.Add(values...)
	return set
}

func (s Set[T]) Add(values ...T) {
	for _, value := range values {
		s[value] = struct{}{}
	}
}

func (s Set[T]) Has(element T) bool {
	_, ok := s[element]

	return ok
}

func (s Set[T]) Values() []T {
	var res []T
	for elem := range s {
		res = append(res, elem)
	}
	return res
}
