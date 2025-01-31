package graphql

import (
	"blog/internal/blog/storage"
	"blog/internal/graphql/model"
	"context"
	"fmt"
	"github.com/gofrs/uuid"
)

type Resolver struct {
	blogDb *storage.Queries
}

func NewResolver(blogDb *storage.Queries) *Resolver {
	return &Resolver{blogDb: blogDb}
}

// CreatePost is the resolver for the createPost field.
func (r *Resolver) CreatePost(ctx context.Context, input model.CreatePostInput) (storage.Post, error) {
	panic(fmt.Errorf("not implemented: CreatePost - createPost"))
}

// PublishPost is the resolver for the publishPost field.
func (r *Resolver) PublishPost(ctx context.Context, id uuid.UUID) (storage.Post, error) {
	panic(fmt.Errorf("not implemented: PublishPost - publishPost"))
}

// DeletePost is the resolver for the deletePost field.
func (r *Resolver) DeletePost(ctx context.Context, id uuid.UUID) (bool, error) {
	panic(fmt.Errorf("not implemented: DeletePost - deletePost"))
}

// Post is the resolver for the post field.
func (r *Resolver) Post(ctx context.Context, id string) (*storage.Post, error) {
	panic(fmt.Errorf("not implemented: Post - post"))
}

// Posts is the resolver for the posts field.
func (r *Resolver) Posts(ctx context.Context) ([]storage.Post, error) {
	return r.blogDb.FindPosts(ctx)
}
