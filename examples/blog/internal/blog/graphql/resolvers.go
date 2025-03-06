package graphql

import (
	"blog/internal/blog/storage"
	"blog/internal/graphql/model"
	"context"
	"fmt"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	// validate input using Ozzo validation
	err := validator.ValidateStructWithContext(
		ctx,
		&input,
		validation.Field(
			&input.Title,
			validation.Required.Error("Title is required"),
		),
		validation.Field(
			&input.Content,
			validation.Required.Error("Content is required"),
		),
	)
	if err != nil {
		return storage.Post{}, err
	}

	preview := input.Content
	if len(input.Content) > 100 {
		preview = input.Content[0:100]
	}

	authorId := auth.GetPerformerID(ctx)
	return r.blogDb.CreatePost(
		ctx, storage.CreatePostParams{
			ID:       uuid.Must(uuid.NewV6()),
			Title:    input.Title,
			Preview:  preview,
			Content:  input.Content,
			AuthorID: authorId,
		},
	)
}

// PublishPost is the resolver for the publishPost field.
func (r *Resolver) PublishPost(ctx context.Context, id uuid.UUID) (storage.Post, error) {
	return r.blogDb.PublishPost(ctx, id)
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
	authorID := auth.GetPerformerID(ctx)
	return r.blogDb.FindPosts(ctx, authorID)
}
