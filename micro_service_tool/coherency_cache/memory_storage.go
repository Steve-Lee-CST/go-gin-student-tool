package coherency_cache

import (
	"context"
	"sync"
	"time"
)

var (
	_ IStorage[any, any] = (*MemoryStorage[any, any])(nil)
)

type MemoryStorage[KT any, VT any] struct {
	BaseStorage[KT]

	cache sync.Map
}

func NewMemoryStorage[KT any, VT any]() *MemoryStorage[KT, VT] {
	return &MemoryStorage[KT, VT]{
		BaseStorage: BaseStorage[KT]{
			lockMap: sync.Map{},
		},
		cache: sync.Map{},
	}
}

func (m *MemoryStorage[KT, VT]) Get(
	ctx context.Context, key *KT, timeout time.Duration,
) (*VT, error) {
	value, ok := m.cache.Load(*key)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if !ok {
		return nil, nil
	}
	return value.(*VT), nil
}

func (m *MemoryStorage[KT, VT]) Set(
	ctx context.Context, key *KT, value *VT, timeout time.Duration,
) error {
	m.cache.Store(*key, value)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (m *MemoryStorage[KT, VT]) Del(
	ctx context.Context, key *KT, timeout time.Duration,
) error {
	m.cache.Delete(*key)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

func (m *MemoryStorage[KT, VT]) Timeout() time.Duration {
	return time.Millisecond * 100
}
