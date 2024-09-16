package errwrap

type Wrapper func(err error) error

func With(w ...Wrapper) Wrapper {
	return func(err error) error {
		if err == nil {
			return nil
		}

		for _, fn := range w {
			err = fn(err)
		}

		return err
	}
}

func Wrap(err error, w ...Wrapper) error {
	if err == nil {
		return nil
	}

	for _, fn := range w {
		err = fn(err)
	}

	return err
}
