// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type CreatePostInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Mutation struct {
}

type PageInfo struct {
	EndCursor       string `json:"endCursor"`
	StartCursor     string `json:"startCursor"`
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
}

type Query struct {
}

type Subscription struct {
}
