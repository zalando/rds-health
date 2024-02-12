//
// Copyright (c) 2024 Zalando SE
//
// This file may be modified and distributed under the terms
// of the MIT license. See the LICENSE file for details.
// https://github.com/zalando/rds-health
//

package cache

import "context"

type Getter[K comparable, V any] interface {
	Lookup(context.Context, K) (V, error)
}

type Cache[K comparable, V any] struct {
	keyval map[K]V
	getter Getter[K, V]
}

func New[K comparable, V any](getter Getter[K, V]) *Cache[K, V] {
	return &Cache[K, V]{
		keyval: make(map[K]V),
		getter: getter,
	}

}

func (c Cache[K, V]) Lookup(ctx context.Context, key K) (V, error) {
	if v, has := c.keyval[key]; has {
		return v, nil
	}

	v, err := c.getter.Lookup(ctx, key)
	if err != nil {
		return *new(V), nil
	}

	c.keyval[key] = v

	return v, nil
}
