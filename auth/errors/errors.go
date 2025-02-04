package errors

import "github.com/go-modulus/modulus/errors"

var ErrIdentityExists = errors.New("identity exists")
var ErrIdentityNotFound = errors.New("identity not found")
var ErrCannotCreateIdentity = errors.New("cannot create identity")
var ErrCannotHashPassword = errors.New("cannot hash password")
var ErrCannotCreateCredential = errors.New("cannot create credential")
