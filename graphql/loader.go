package graphql

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/vektah/gqlparser/v2/ast"
	"sync"
)

type contextKey string

const (
	LoadersKey contextKey = "Loaders"
)

type LoaderFactory[K comparable, T any] interface {
	Create() *dataloader.Loader[K, T]
}

func WithLoaders(ctx context.Context, loaders *sync.Map) context.Context {
	return context.WithValue(ctx, LoadersKey, loaders)
}

func getLoaders(ctx context.Context) *sync.Map {
	if loaders, ok := ctx.Value(LoadersKey).(*sync.Map); ok {
		return loaders
	}
	return nil
}

func GetLoader[K comparable, T any](ctx context.Context, factory LoaderFactory[K, T]) *dataloader.Loader[K, T] {
	loaders := getLoaders(ctx)
	key := fmt.Sprintf("%T", factory)
	existLoader, ok := loaders.Load(key)
	if ok {
		if loader, ok := existLoader.(*dataloader.Loader[K, T]); ok {
			return loader
		}

		var (
			loaderKey   K
			loaderValue T
		)
		panic(fmt.Sprintf("loader with key %s is not of type *dataloader.Loader[%T, %T]", key, loaderKey, loaderValue))
	}

	loader := factory.Create()
	loaders.Store(key, loader)
	return loader
}

type LoadersInitializer struct{}

func NewLoadersInitializer() *LoadersInitializer {
	return &LoadersInitializer{}
}

var _ interface {
	graphql.HandlerExtension
	graphql.OperationInterceptor
	graphql.ResponseInterceptor
} = LoadersInitializer{}

func (LoadersInitializer) ExtensionName() string {
	return "LoadersInitializer"
}

func (LoadersInitializer) Validate(graphql.ExecutableSchema) error {
	return nil
}

// InterceptOperation Init loaders before each operation
func (l LoadersInitializer) InterceptOperation(
	ctx context.Context,
	next graphql.OperationHandler,
) graphql.ResponseHandler {
	return next(WithLoaders(ctx, &sync.Map{}))
}

// InterceptResponse Init loaders before each event in subscription
func (l LoadersInitializer) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	oc := graphql.GetOperationContext(ctx)
	operation := oc.Operation

	if operation == nil {
		return next(ctx)
	}

	if operation.Operation == ast.Subscription {
		ctx = WithLoaders(ctx, &sync.Map{})
	}

	return next(ctx)
}
