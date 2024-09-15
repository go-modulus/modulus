package utils

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

func First[T any](v T, _ ...any) T {
	return v
}
