package utils

import (
	"cmp"
	"fmt"
	"slices"
)

func MergeSlices[T comparable](chunks [][]T) []T {
	results := make([]T, 0)
	for _, chunk := range chunks {
		// 保证至少能容纳当前 chunk 长度的额外容量
		// Grow 保留了原切片的长度不变
		results = slices.Grow(results, len(chunk))
		results = append(results, chunk...)
	}
	return results
}

func DistinctSlice[T comparable](s []T) []T {
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

func DistinctSliceV2[T cmp.Ordered](s []T) []T {
	slices.Sort(s)
	// 原地消除连续重复元素, 返回有效长度截断后的新切片
	return slices.Compact(s)
}

// SliceToMap returns map for slice of structs by spec key.
func SliceToMap[K comparable, V any](s []V, keyFunc func(V) K) map[K]V {
	m := make(map[K]V, len(s))
	for _, v := range s {
		m[keyFunc(v)] = v
	}
	return m
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
