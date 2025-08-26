package errtrace

import "github.com/go-modulus/modulus/errors"

func Wrap(err error) error {
	if err == nil {
		return nil
	}
	return errors.WithTraceSkip(err, 2)
}

// Wrap2 is used to [Wrap] the last error return when returning 2 values.
// This is useful when returning multiple returns from a function call directly:
//
//	return Wrap2(fn())
//
// Wrap2 is used by the CLI to avoid line number changes.
//
//go:noinline due to GetCaller (see [Wrap] for details).
func Wrap2[T any](t T, err error) (T, error) {
	if err == nil {
		return t, nil
	}

	return t, errors.WithTraceSkip(err, 2)
}

// Wrap3 is used to [Wrap] the last error return when returning 3 values.
// This is useful when returning multiple returns from a function call directly:
//
//	return Wrap3(fn())
//
// Wrap3 is used by the CLI to avoid line number changes.
//
//go:noinline due to GetCaller (see [Wrap] for details).
func Wrap3[T1, T2 any](t1 T1, t2 T2, err error) (T1, T2, error) {
	if err == nil {
		return t1, t2, nil
	}

	return t1, t2, errors.WithTraceSkip(err, 2)
}

// Wrap4 is used to [Wrap] the last error return when returning 4 values.
// This is useful when returning multiple returns from a function call directly:
//
//	return Wrap4(fn())
//
// Wrap4 is used by the CLI to avoid line number changes.
//
//go:noinline due to GetCaller (see [Wrap] for details).
func Wrap4[T1, T2, T3 any](t1 T1, t2 T2, t3 T3, err error) (T1, T2, T3, error) {
	if err == nil {
		return t1, t2, t3, nil
	}

	return t1, t2, t3, errors.WithTraceSkip(err, 2)
}

// Wrap5 is used to [Wrap] the last error return when returning 5 values.
// This is useful when returning multiple returns from a function call directly:
//
//	return Wrap5(fn())
//
// Wrap5 is used by the CLI to avoid line number changes.
//
//go:noinline due to GetCaller (see [Wrap] for details).
func Wrap5[T1, T2, T3, T4 any](t1 T1, t2 T2, t3 T3, t4 T4, err error) (T1, T2, T3, T4, error) {
	if err == nil {
		return t1, t2, t3, t4, nil
	}

	return t1, t2, t3, t4, errors.WithTraceSkip(err, 2)
}

// Wrap6 is used to [Wrap] the last error return when returning 6 values.
// This is useful when returning multiple returns from a function call directly:
//
//	return Wrap6(fn())
//
// Wrap6 is used by the CLI to avoid line number changes.
//
//go:noinline due to GetCaller (see [Wrap] for details).
func Wrap6[T1, T2, T3, T4, T5 any](t1 T1, t2 T2, t3 T3, t4 T4, t5 T5, err error) (T1, T2, T3, T4, T5, error) {
	if err == nil {
		return t1, t2, t3, t4, t5, nil
	}

	return t1, t2, t3, t4, t5, errors.WithTraceSkip(err, 2)
}
