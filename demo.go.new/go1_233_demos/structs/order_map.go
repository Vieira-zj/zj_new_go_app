package structs

import (
	"iter"
	"slices"
	"strings"
)

type OrderMap[T int | string] struct {
	orderValues []T
	valuesMap   map[T]any
}

func (m *OrderMap[T]) CreateBySlice(list []T) {
	m.orderValues = list
	for _, v := range list {
		m.valuesMap[v] = struct{}{}
	}
}

func (m *OrderMap[T]) CreateByMap(items map[T]any) {
	m.valuesMap = items
	m.orderValues = make([]T, 0, len(items))
	for k := range items {
		m.orderValues = append(m.orderValues, k)
	}
}

func (m *OrderMap[T]) Slice() []T {
	return m.orderValues
}

func (m *OrderMap[T]) Put(key T, value any) {
	m.valuesMap[key] = value
}

func (m *OrderMap[T]) Get(key T) (any, bool) {
	val, ok := m.valuesMap[key]
	return val, ok
}

func (m *OrderMap[T]) Sort() {
	slices.SortFunc(m.orderValues, func(a, b T) int {
		inta, oka := any(a).(int)
		intb, okb := any(b).(int)
		if oka && okb {
			return inta - intb
		}
		stra, oka := any(a).(string)
		strb, okb := any(b).(string)
		if oka && okb {
			return strings.Compare(stra, strb)
		}
		return 1
	})
}

func (m *OrderMap[T]) SeqValues() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range m.orderValues {
			if !yield(v) {
				break
			}
		}
	}
}

func (m *OrderMap[T]) SeqKeyValues() iter.Seq2[T, any] {
	return func(yield func(T, any) bool) {
		for k, v := range m.valuesMap {
			if !yield(k, v) {
				break
			}
		}
	}
}

func (m *OrderMap[T]) StableSeqKeyValues() iter.Seq2[T, any] {
	return func(yield func(T, any) bool) {
		for _, k := range m.orderValues {
			v := m.valuesMap[k]
			if !yield(k, v) {
				break
			}
		}
	}
}
