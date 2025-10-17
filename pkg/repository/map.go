package repository

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrNotFound     = errors.New("repository: not found")
	ErrTypeMismatch = errors.New("repository: type mismatch")
)

type SyncRepository[V comparable, T Entity[V]] struct {
	m sync.Map
}

func NewSyncRepository[V comparable, T Entity[V]]() Repository[V, T] {
	return &SyncRepository[V, T]{}
}

func (r *SyncRepository[V, T]) Insert(ctx context.Context, t T) (T, error) {
	r.m.Store(t.Id(), t)
	return t, nil
}

func (r *SyncRepository[V, T]) Find(ctx context.Context, v V) (T, error) {
	val, ok := r.m.Load(v)
	if !ok {
		var zero T
		return zero, ErrNotFound
	}
	t, ok := val.(T)
	if !ok {
		return t, ErrTypeMismatch
	}

	return t, nil
}

func (r *SyncRepository[V, T]) Update(ctx context.Context, t T) (T, error) {
	r.m.Store(t.Id(), t)
	return t, nil
}

func (r *SyncRepository[V, T]) Delete(ctx context.Context, v V) error {
	r.m.Delete(v)
	return nil
}

func (r *SyncRepository[V, T]) FindList(ctx context.Context) ([]T, error) {
	var list []T
	r.m.Range(func(_, t interface{}) bool {
		if id, ok := t.(T); ok {
			list = append(list, id)
		}
		return true
	})
	return list, nil
}
