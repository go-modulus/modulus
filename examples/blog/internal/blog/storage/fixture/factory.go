package fixture

import (
	"blog/internal/blog/storage"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Factory struct {
	db storage.DBTX
}

func NewFactory(db storage.DBTX) *Factory {
	return &Factory{
		db: db,
	}
}

func (f *Factory) NewPostFixture() *PostFixture {
	id := uuid.Must(uuid.NewV6())
	return NewPostFixture(
		f.db, storage.Post{
			ID:          id,
			Title:       "Title " + id.String(),
			Preview:     "Preview " + id.String(),
			Content:     "Content " + id.String(),
			Status:      storage.PostStatusPublished,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PublishedAt: null.TimeFrom(time.Now()),
			DeletedAt:   null.Time{},
			AuthorID:    uuid.Must(uuid.NewV6()),
		},
	)
}
