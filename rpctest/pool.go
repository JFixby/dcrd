// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpctest

import (
	"fmt"
	"sync"
)

// Pool keeps track of reusable Spawnable instances.
// Multiple Spawnable instances may be run concurrently, to allow for testing
// complex scenarios involving multiple nodes.
type Pool struct {
	cache                    map[string]Spawnable
	registryAccessController sync.RWMutex
	spawner                  Spawner
}

func NewPool(spawner Spawner) *Pool {
	return &Pool{
		cache:   make(map[string]Spawnable),
		spawner: spawner,
	}
}

type Spawnable interface {
}

// Spawner manages new Spawnable instance creation and disposal
type Spawner interface {
	// NewInstance must return a new freshly created Spawnable instance
	NewInstance(spawnableName string) Spawnable

	// NameForTag defines a policy for mapping input tags to Spawnable names
	NameForTag(tag string) string

	// Dispose should take care of Spawnable instance disposal
	Dispose(spawnableToDispose Spawnable) error
}

// ObtainSpawnable returns reusable Spawnable instance upon request,
// creates a new instance when required and stores it in the cache
// for the following calls
func (pool *Pool) ObtainSpawnable(tag string) Spawnable {

	// Resolve Spawnable name for the tag requested
	// Pool uses SpawnableName as a key to cache a Spawnable instance
	spawnableName := pool.spawner.NameForTag(tag)

	spawnable := pool.cache[spawnableName]

	// Create and cache a new instance when not present in cache
	if spawnable == nil {
		spawnable = pool.spawner.NewInstance(spawnableName)
		pool.cache[spawnableName] = spawnable
	}

	return spawnable
}

// ObtainSpawnableConcurrentSafe is safe for concurrent access.
func (pool *Pool) ObtainSpawnableConcurrentSafe(tag string) Spawnable {
	pool.registryAccessController.Lock()
	defer pool.registryAccessController.Unlock()
	return pool.ObtainSpawnable(tag)
}

// TearDownAll disposes all instances in the cache
func (pool *Pool) TearDownAll() {
	for key, spawnable := range pool.cache {
		err := pool.spawner.Dispose(spawnable)
		delete(pool.cache, key)
		if err != nil {
			fmt.Printf("Failed to dispose Spawnable <%v>: %v", key, err)
		}
	}
}

// InitTags ensures the cache will immediately resolve tags from the given list
func (pool *Pool) InitTags(tags []string) {
	for _, tag := range tags {
		pool.ObtainSpawnable(tag)
	}
}

func (pool *Pool) Size() int {
	return len(pool.cache)
}
