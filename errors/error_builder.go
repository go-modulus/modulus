package errors

import "errors"

type Builder struct {
	err error
}

func NewB(err string) *Builder {
	// it is a hack to mark the error for extracting to the translation file
	//_ = ht.Sprintf(err)
	return &Builder{err: WrapHint(errors.New(err), err)}
}

func NewBE(err error) *Builder {
	return &Builder{err: err}
}

func (b *Builder) WithTags(tags ...string) *Builder {
	b.err = WrapAddingTags(b.err, tags...)
	return b
}

func (b *Builder) WithHint(hint string) *Builder {
	b.err = WrapHint(b.err, hint)
	return b
}

func (b *Builder) WithCause(cause error) *Builder {
	b.err = WrapCause(b.err, cause)
	return b
}

func (b *Builder) WithMeta(kv ...string) *Builder {
	b.err = WrapMeta(b.err, kv...)
	return b
}

func (b *Builder) LogAsError() *Builder {
	b.err = WrapAddingTags(b.err, LogAsError)
	return b
}

func (b *Builder) LogAsWarning() *Builder {
	b.err = WrapAddingTags(b.err, LogAsWarn)
	return b
}

func (b *Builder) LogAsInfo() *Builder {
	b.err = WrapAddingTags(b.err, LogAsInfo)
	return b
}

func (b *Builder) LogAsDebug() *Builder {
	b.err = WrapAddingTags(b.err, LogAsDebug)
	return b
}

func (b *Builder) Build() error {
	return b.err
}
