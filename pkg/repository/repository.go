package repository

import (
	"context"
)

type Entity[V comparable] interface {
	Id() V
}

type Repository[V comparable, T Entity[V]] interface {
	Insert(ctx context.Context, t T) (T, error)
	Find(ctx context.Context, v V) (T, error)
	Delete(ctx context.Context, v V) error
	FindList(ctx context.Context) ([]T, error)
}
