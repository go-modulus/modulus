package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"blog/internal/blog/storage"
	"blog/internal/graphql/model"
	storage1 "blog/internal/user/storage"
	"context"
	"github.com/gofrs/uuid"
)

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (storage.Post, error) {
	return r.blogResolver.CreatePost(ctx, input)
}

// PublishPost is the resolver for the publishPost field.
func (r *mutationResolver) PublishPost(ctx context.Context, id uuid.UUID) (storage.Post, error) {
	return r.blogResolver.PublishPost(ctx, id)
}

// DeletePost is the resolver for the deletePost field.
func (r *mutationResolver) DeletePost(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.blogResolver.DeletePost(ctx, id)
}

// Author is the resolver for the author field.
func (r *postResolver) Author(ctx context.Context, obj *storage.Post) (storage1.User, error) {
	return r.userLoaderFactory.UserLoader().Load(ctx, obj.AuthorID)
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*storage.Post, error) {
	return r.blogResolver.Post(ctx, id)
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context) ([]storage.Post, error) {
	return r.blogResolver.Posts(ctx)
}
