package coherency_cache

import (
	"context"
	"sync"
	"time"
)

var (
	_ IStorage[any, any] = (*CoherencyStorage[any, any])(nil)
)

type IStorage[KT any, VT any] interface {
	Get(ctx context.Context, key *KT, timeout time.Duration) (*VT, error)
	Set(ctx context.Context, key *KT, value *VT, timeout time.Duration) error
	Del(ctx context.Context, key *KT, timeout time.Duration) error

	Lock(ctx context.Context, key *KT)
	Unlock(ctx context.Context, key *KT)

	Timeout() time.Duration
}

type BaseStorage[KT any] struct {
	lockMap sync.Map
}

func (b *BaseStorage[KT]) Lock(ctx context.Context, key *KT) {
	lock, _ := b.lockMap.LoadOrStore(*key, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
}

func (b *BaseStorage[KT]) Unlock(ctx context.Context, key *KT) {
	lock, ok := b.lockMap.Load(*key)
	if !ok {
		return
	}
	lock.(*sync.Mutex).Unlock()
}

type CoherencyStorage[KT any, VT any] struct {
	BaseStorage[KT]

	cache  IStorage[KT, VT]
	source IStorage[KT, VT]
}

func NewCoherencyStorage[KT any, VT any](
	cache IStorage[KT, VT], source IStorage[KT, VT],
) *CoherencyStorage[KT, VT] {
	return &CoherencyStorage[KT, VT]{
		cache:  cache,
		source: source,
		BaseStorage: BaseStorage[KT]{
			lockMap: sync.Map{},
		},
	}
}

func (c *CoherencyStorage[KT, VT]) Get(
	ctx context.Context, key *KT, timeout time.Duration,
) (*VT, error) {
	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cacheValue, err := c.cache.Get(subCtx, key, c.cache.Timeout())
	if err != nil || cacheValue != nil {
		return cacheValue, err
	}

	c.cache.Lock(subCtx, key)
	defer c.cache.Unlock(subCtx, key)

	cacheValue, err = c.source.Get(subCtx, key, c.cache.Timeout())
	if err != nil || cacheValue != nil {
		return cacheValue, err
	}

	sourceValue, err := c.source.Get(subCtx, key, c.cache.Timeout())
	if err != nil {
		return nil, err
	}

	if err = c.cache.Set(subCtx, key, sourceValue, c.cache.Timeout()); err != nil {
		return nil, err
	}

	return sourceValue, nil
}

func (c *CoherencyStorage[KT, VT]) Set(
	ctx context.Context, key *KT, value *VT, timeout time.Duration,
) error {
	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := func() error {
		c.source.Lock(subCtx, key)
		defer c.source.Unlock(subCtx, key)

		if err := c.source.Set(subCtx, key, value, c.source.Timeout()); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		c.cache.Lock(subCtx, key)
		defer c.cache.Unlock(subCtx, key)

		if err := c.cache.Set(subCtx, key, value, c.cache.Timeout()); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func (c *CoherencyStorage[KT, VT]) Del(
	ctx context.Context, key *KT, timeout time.Duration,
) error {
	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := func() error {
		c.source.Lock(subCtx, key)
		defer c.source.Unlock(subCtx, key)

		if err := c.source.Del(subCtx, key, c.source.Timeout()); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		c.cache.Lock(subCtx, key)
		defer c.cache.Unlock(subCtx, key)

		if err := c.cache.Del(subCtx, key, c.cache.Timeout()); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func (c *CoherencyStorage[KT, VT]) Lock(
	_ context.Context, key *KT,
) {
	lock, _ := c.lockMap.LoadOrStore(*key, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
}

func (c *CoherencyStorage[KT, VT]) Unlock(
	_ context.Context, key *KT,
) {
	lock, ok := c.lockMap.Load(*key)
	if !ok {
		return
	}
	lock.(*sync.Mutex).Unlock()
}

func (c *CoherencyStorage[KT, VT]) Timeout() time.Duration {
	return c.cache.Timeout()
}
