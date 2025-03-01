package graphql_test

import (
	"blog/internal/graphql/model"
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResolver_CreatePost(t *testing.T) {
	t.Parallel()
	t.Run(
		"create post", func(t *testing.T) {
			t.Parallel()
			ctx := auth.WithPerformer(
				context.Background(),
				auth.Performer{
					ID: uuid.Must(uuid.NewV6()),
				},
			)
			post, err := resolver.CreatePost(
				ctx, model.CreatePostInput{
					Title:   "Title",
					Content: "Content",
				},
			)

			savedPost := fixtures.NewPostFixture().ID(post.ID).PullUpdates(t).Cleanup(t).GetEntity()

			t.Log("When the post is created with valid input")
			t.Log("	Then the post should be created successfully")
			require.NoError(t, err)
			require.NotEqual(t, uuid.Nil, post.ID)
			require.Equal(t, "Title", post.Title)
			require.Equal(t, "Content", post.Content)
			t.Log("	And the post should be saved in the database")
			require.Equal(t, post.ID, savedPost.ID)
			require.Equal(t, post.Title, savedPost.Title)
			require.Equal(t, post.Content, savedPost.Content)
		},
	)
}
