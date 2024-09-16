package errbuilder

import (
	"errors"
	errors2 "github.com/go-modulus/modulus/errors"
	"github.com/go-modulus/modulus/errors/errlog"
)

type Builder struct {
	err error
}

func New(err string) *Builder {
	// it is a hack to mark the error for extracting to the translation file
	//_ = ht.Sprintf(err)
	return &Builder{err: errors2.WrapHint(errors.New(err), err)}
}

func NewE(err error) *Builder {
	return &Builder{err: err}
}

func (b *Builder) WithTags(tags ...string) *Builder {
	b.err = errors2.WrapAddingTags(b.err, tags...)
	return b
}

func (b *Builder) WithHint(hint string) *Builder {
	b.err = errors2.WrapHint(b.err, hint)
	return b
}

func (b *Builder) WithCause(cause error) *Builder {
	b.err = errors2.WrapCause(b.err, cause)
	return b
}

func (b *Builder) WithMeta(kv ...string) *Builder {
	b.err = errors2.WrapMeta(b.err, kv...)
	return b
}

func (b *Builder) LogAsError() *Builder {
	b.err = errors2.WrapAddingTags(b.err, errlog.LogAsError)
	return b
}

func (b *Builder) LogAsWarning() *Builder {
	b.err = errors2.WrapAddingTags(b.err, errlog.LogAsWarn)
	return b
}

func (b *Builder) LogAsInfo() *Builder {
	b.err = errors2.WrapAddingTags(b.err, errlog.LogAsInfo)
	return b
}

func (b *Builder) LogAsDebug() *Builder {
	b.err = errors2.WrapAddingTags(b.err, errlog.LogAsDebug)
	return b
}

func (b *Builder) Build() error {
	return b.err
}
