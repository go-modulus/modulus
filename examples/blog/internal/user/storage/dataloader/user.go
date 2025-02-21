package dataloader

import (
	"blog/internal/user/storage"
	"context"
	dl "github.com/debugger84/sqlc-dataloader"
	loaderCache "github.com/debugger84/sqlc-dataloader/cache"
	uuid "github.com/gofrs/uuid"
	"github.com/graph-gophers/dataloader/v7"
	"time"
)

type UserLoader struct {
	innerLoader *dataloader.Loader[uuid.UUID, storage.User]
	db          storage.DBTX
	cache       dataloader.Cache[uuid.UUID, storage.User]
}

func NewUserLoader(
	db storage.DBTX,
	cache dataloader.Cache[uuid.UUID, storage.User],
) *UserLoader {
	if cache == nil {
		ttl, _ := time.ParseDuration("1m")
		cache = loaderCache.NewLRU[uuid.UUID, storage.User](100, ttl)
	}
	return &UserLoader{
		db:    db,
		cache: cache,
	}
}

func (l *UserLoader) getInnerLoader() *dataloader.Loader[uuid.UUID, storage.User] {
	if l.innerLoader == nil {
		l.innerLoader = dataloader.NewBatchedLoader(
			func(ctx context.Context, keys []uuid.UUID) []*dataloader.Result[storage.User] {
				userMap, err := l.findItemsMap(ctx, keys)

				result := make([]*dataloader.Result[storage.User], len(keys))
				for i, key := range keys {
					if err != nil {
						result[i] = &dataloader.Result[storage.User]{Data: storage.User{}, Error: err}
						continue
					}

					if loadedItem, ok := userMap[key]; ok {
						result[i] = &dataloader.Result[storage.User]{Data: loadedItem}
					} else {
						result[i] = &dataloader.Result[storage.User]{Data: storage.User{}, Error: dl.ErrNoRows}
					}
				}
				return result
			},
			dataloader.WithCache(l.cache),
		)
	}
	return l.innerLoader
}

func (l *UserLoader) findItemsMap(ctx context.Context, keys []uuid.UUID) (map[uuid.UUID]storage.User, error) {
	res := make(map[uuid.UUID]storage.User, len(keys))

	query := `SELECT id, email, name, created_at, updated_at FROM "user"."user" WHERE id = ANY($1)`
	rows, err := l.db.Query(ctx, query, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var result storage.User
		err := rows.Scan(
			&result.ID,
			&result.Email,
			&result.Name,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		res[result.ID] = result
	}
	return res, nil
}

func (l *UserLoader) Load(ctx context.Context, userKey uuid.UUID) (storage.User, error) {
	return l.getInnerLoader().Load(ctx, userKey)()
}
