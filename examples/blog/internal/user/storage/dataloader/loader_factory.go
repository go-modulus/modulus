package dataloader

import (
	"blog/internal/user/storage"
)

type LoaderFactory struct {
	db         storage.DBTX
	userLoader *UserLoader
}

func NewLoaderFactory(db storage.DBTX) *LoaderFactory {
	return &LoaderFactory{
		db: db,
	}
}

func (f *LoaderFactory) UserLoader() *UserLoader {
	if f.userLoader == nil {
		f.userLoader = NewUserLoader(f.db, nil)
	}
	return f.userLoader
}
