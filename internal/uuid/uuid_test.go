// Copyright (C) MongoDB, Inc. 2022-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package uuid

import (
	crand "crypto/rand"
	mrand "math/rand"
	"sync"
	"testing"

	"github.com/seehuhn/mt19937"
	xrand "golang.org/x/exp/rand"

	"go.mongodb.org/mongo-driver/internal/randutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type randReader struct{}

func (r *randReader) Read(b []byte) (n int, err error) {
	return crand.Read(b)
}

type LockedRandUUID struct {
	mu sync.Mutex
}

func (l *LockedRandUUID) new() (uuid.UUID, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return uuid.NewRandom()
}

func TestNew(t *testing.T) {
	m := make(map[UUID]bool)
	for i := 1; i < 1000000; i++ {
		uuid, err := New()
		require.NoError(t, err, "New error")
		require.False(t, m[uuid], "New returned a duplicate UUID %v", uuid)
		m[uuid] = true
	}
}

// GODRIVER-2349
// Test that initializing many package-global UUID sources concurrently never leads to any duplicate
// UUIDs being generated.
func TestGlobalSource(t *testing.T) {
	t.Run("math rand", func(t *testing.T) {
		// Create a slice of 1,000 sources and initialize them in 1,000 separate goroutines. The goal is
		// to emulate many separate Go driver processes starting at the same time and initializing the
		// uuid package at the same time.
		sources := make([]*source, 1000)
		var wg sync.WaitGroup
		for i := range sources {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sources[i] = newGlobalSource(randutil.NewLockedRand(mrand.NewSource(randutil.CryptoSeed())))
			}(i)
		}
		wg.Wait()

		// Read 1,000 UUIDs from each source and assert that there is never a duplicate value, either
		// from the same source or from separate sources.
		const iterations = 1000
		uuids := make(map[UUID]bool, len(sources)*iterations)
		for i := 0; i < iterations; i++ {
			for j, s := range sources {
				uuid, err := s.new()
				require.NoError(t, err, "new() error")
				require.Falsef(t, uuids[uuid], "source %d returned a duplicate UUID on iteration %d: %v", j, i, uuid)
				uuids[uuid] = true
			}
		}
	})
	t.Run("crypto rand", func(t *testing.T) {
		// Create a slice of 1,000 sources and initialize them in 1,000 separate goroutines. The goal is
		// to emulate many separate Go driver processes starting at the same time and initializing the
		// uuid package at the same time.
		sources := make([]*source, 1000)
		var wg sync.WaitGroup
		for i := range sources {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sources[i] = newGlobalSource(new(randReader))
			}(i)
		}
		wg.Wait()

		// Read 1,000 UUIDs from each source and assert that there is never a duplicate value, either
		// from the same source or from separate sources.
		const iterations = 1000
		uuids := make(map[UUID]bool, len(sources)*iterations)
		for i := 0; i < iterations; i++ {
			for j, s := range sources {
				uuid, err := s.new()
				require.NoError(t, err, "new() error")
				require.Falsef(t, uuids[uuid], "source %d returned a duplicate UUID on iteration %d: %v", j, i, uuid)
				uuids[uuid] = true
			}
		}
	})
	t.Run("exp rand", func(t *testing.T) {
		// Create a slice of 1,000 sources and initialize them in 1,000 separate goroutines. The goal is
		// to emulate many separate Go driver processes starting at the same time and initializing the
		// uuid package at the same time.
		sources := make([]*source, 1000)
		var wg sync.WaitGroup
		for i := range sources {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sources[i] = newGlobalSource(xrand.New(xrand.NewSource((uint64)(randutil.CryptoSeed()))))
			}(i)
		}
		wg.Wait()

		// Read 1,000 UUIDs from each source and assert that there is never a duplicate value, either
		// from the same source or from separate sources.
		const iterations = 1000
		uuids := make(map[UUID]bool, len(sources)*iterations)
		for i := 0; i < iterations; i++ {
			for j, s := range sources {
				uuid, err := s.new()
				require.NoError(t, err, "new() error")
				require.Falsef(t, uuids[uuid], "source %d returned a duplicate UUID on iteration %d: %v", j, i, uuid)
				uuids[uuid] = true
			}
		}
	})
	t.Run("mt19937", func(t *testing.T) {
		// Create a slice of 1,000 sources and initialize them in 1,000 separate goroutines. The goal is
		// to emulate many separate Go driver processes starting at the same time and initializing the
		// uuid package at the same time.
		sources := make([]*source, 1000)
		var wg sync.WaitGroup
		for i := range sources {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ins := mt19937.New()
				ins.Seed(randutil.CryptoSeed())
				sources[i] = newGlobalSource(ins)
			}(i)
		}
		wg.Wait()

		// Read 1,000 UUIDs from each source and assert that there is never a duplicate value, either
		// from the same source or from separate sources.
		const iterations = 1000
		uuids := make(map[UUID]bool, len(sources)*iterations)
		for i := 0; i < iterations; i++ {
			for j, s := range sources {
				uuid, err := s.new()
				require.NoError(t, err, "new() error")
				require.Falsef(t, uuids[uuid], "source %d returned a duplicate UUID on iteration %d: %v", j, i, uuid)
				uuids[uuid] = true
			}
		}
	})
	t.Run("uuid", func(t *testing.T) {
		// Read 1,000,000 UUIDs from each source and assert that there is never a duplicate value, either
		// from the same source or from separate sources.
		const iterations = 1000 * 1000
		uuids := make(map[uuid.UUID]bool, iterations)
		for i := 0; i < iterations; i++ {
			uuid, err := uuid.NewRandom()
			require.NoError(t, err, "uuid.NewRandom() error")
			require.Falsef(t, uuids[uuid], "returned a duplicate UUID on iteration %d: %v", i, uuid)
			uuids[uuid] = true
		}
	})
	t.Run("pooled uuid", func(t *testing.T) {
		// Read 1,000,000 UUIDs from each source and assert that there is never a duplicate value, either
		// from the same source or from separate sources.
		uuid.EnableRandPool()
		defer uuid.DisableRandPool()
		var s LockedRandUUID
		const iterations = 1000 * 1000
		uuids := make(map[uuid.UUID]bool, iterations)
		for i := 0; i < iterations; i++ {
			uuid, err := s.new()
			require.NoError(t, err, "uuid.NewRandom() error")
			require.Falsef(t, uuids[uuid], "returned a duplicate UUID on iteration %d: %v", i, uuid)
			uuids[uuid] = true
		}
	})
}

func BenchmarkUuidGeneration(b *testing.B) {
	b.Run("math rand", func(b *testing.B) {
		s := newGlobalSource(randutil.NewLockedRand(mrand.NewSource(randutil.CryptoSeed())))
		for i := 0; i < b.N; i++ {
			s.new()
		}
	})
	b.Run("crypto rand", func(b *testing.B) {
		s := newGlobalSource(new(randReader))
		for i := 0; i < b.N; i++ {
			s.new()
		}
	})
	b.Run("exp rand", func(b *testing.B) {
		s := newGlobalSource(xrand.New(xrand.NewSource((uint64)(randutil.CryptoSeed()))))
		for i := 0; i < b.N; i++ {
			s.new()
		}
	})
	b.Run("mt19937 rand", func(b *testing.B) {
		ins := mt19937.New()
		ins.Seed(randutil.CryptoSeed())
		s := newGlobalSource(ins)
		for i := 0; i < b.N; i++ {
			s.new()
		}
	})
	b.Run("uuid", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			uuid.NewRandom()
		}
	})
	b.Run("pooled uuid", func(b *testing.B) {
		uuid.EnableRandPool()
		defer uuid.DisableRandPool()
		var s LockedRandUUID
		for i := 0; i < b.N; i++ {
			s.new()
		}
	})
}
