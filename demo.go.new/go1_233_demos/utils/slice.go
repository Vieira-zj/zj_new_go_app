package utils

import (
	"fmt"
	"slices"
)

func SliceDistinct[T comparable](s []T) []T {
	if len(s) < 2 {
		return s
	}

	m := make(map[T]struct{})
	results := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			results = append(results, v)
		}
	}
	return results
}

// Differences between two slices, including added, removed, and matched elements.

type SliceDiffs[T comparable] struct {
	Added   []T
	Removed []T
	Matched []T
}

func NewSliceDiff[T comparable]() SliceDiffs[T] {
	return SliceDiffs[T]{
		Added:   []T{},
		Removed: []T{},
		Matched: []T{},
	}
}

func (s SliceDiffs[T]) Equal(other SliceDiffs[T]) bool {
	if !slices.Equal(s.Added, other.Added) {
		return false
	}
	if !slices.Equal(s.Removed, other.Removed) {
		return false
	}
	if !slices.Equal(s.Matched, other.Matched) {
		return false
	}
	return true
}

func (s SliceDiffs[T]) String() string {
	return fmt.Sprintf("Added: %v, Removed: %v, Matched: %v", s.Added, s.Removed, s.Matched)
}

func SliceDiff[T comparable](s1, s2 []T) SliceDiffs[T] {
	m := make(map[T]struct{}, len(s1))
	for _, v := range s1 {
		m[v] = struct{}{}
	}

	diffs := NewSliceDiff[T]()
	for _, v := range s2 {
		if _, ok := m[v]; !ok {
			diffs.Added = append(diffs.Added, v)
		} else {
			diffs.Matched = append(diffs.Matched, v)
		}
		delete(m, v)
	}

	for v := range m {
		diffs.Removed = append(diffs.Removed, v)
	}
	return diffs
}
