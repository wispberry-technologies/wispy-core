// Source: https://github.com/eliben/gogl/blob/main/hashset/hashset.go
// Created by: Eli Bendersky (https://eli.thegreenplace.net)
// License: "The Unlicense" (https://unlicense.org/) source:(https://github.com/eliben/gogl/blob/main/LICENSE)

// Package hashset provides a map-based Set.
package common

import "iter"

// HashSet is a generic set based on a hash table (map).
type HashSet[T comparable] struct {
	m map[T]struct{}
}

// New creates a new HashSet.
func NewHashSet[T comparable]() *HashSet[T] {
	return &HashSet[T]{m: make(map[T]struct{})}
}

// InitWith creates a new HashSet initialized with vals.
func InitWith[T comparable](vals ...T) *HashSet[T] {
	hs := NewHashSet[T]()
	for _, v := range vals {
		hs.Add(v)
	}
	return hs
}

// Add adds a value to the set.
func (hs *HashSet[T]) Add(val T) {
	hs.m[val] = struct{}{}
}

// Contains reports whether the set contains the given value.
func (hs *HashSet[T]) Contains(val T) bool {
	_, ok := hs.m[val]
	return ok
}

// Len returns the size/length of the set - the number of values it contains.
func (hs *HashSet[T]) Len() int {
	return len(hs.m)
}

// Delete removes a value from the set; if the value doesn't exist in the
// set, this is a no-op.
func (hs *HashSet[T]) Delete(val T) {
	delete(hs.m, val)
}

// All returns an iterator over all the values in the set.
func (hs *HashSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range hs.m {
			if !yield(v) {
				return
			}
		}
	}
}

// Union returns the set union of hs with other. It creates a new set.
func (hs *HashSet[T]) Union(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range hs.m {
		result.Add(v)
	}
	for v := range other.m {
		result.Add(v)
	}
	return result
}

// Intersection returns the set intersection of hs with other. It creates a
// new set.
func (hs *HashSet[T]) Intersection(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range hs.m {
		if other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns the set difference hs - other. It creates a new set.
func (hs *HashSet[T]) Difference(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range hs.m {
		if !other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}
