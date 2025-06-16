package utils

// Set defines a set by map. Not concurrent security.
// TODO: 优化参考 "k8s.io/apimachinery/pkg/util/sets"
type Set struct {
	data map[interface{}]interface{}
}

// NewSet creates a instance of set.
func NewSet(size int, items ...interface{}) *Set {
	set := &Set{
		data: make(map[interface{}]interface{}, size),
	}
	for _, item := range items {
		set.Add(item)
	}
	return set
}

// Len .
func (s *Set) Len() int {
	return len(s.data)
}

// Add .
func (s *Set) Add(val interface{}) {
	if _, ok := s.data[val]; !ok {
		s.data[val] = nil
	}
}

// Remove .
func (s *Set) Remove(val interface{}) {
	delete(s.data, val)
}

// Has .
func (s *Set) Has(val interface{}) bool {
	_, ok := s.data[val]
	return ok
}

// ToSlice .
func (s *Set) ToSlice() []interface{} {
	ret := make([]interface{}, 0, s.Len())
	s.ForEach(func(item interface{}) bool {
		ret = append(ret, item)
		return true
	})
	return ret
}

// ForEach .
func (s *Set) ForEach(cb func(item interface{}) bool) {
	for k := range s.data {
		if !cb(k) {
			return
		}
	}
}

// Intersect .
func (s *Set) Intersect(another *Set) *Set {
	len := s.Len()
	if another.Len() < s.Len() {
		len = another.Len()
	}
	ret := NewSet(len)

	s.ForEach(func(item interface{}) bool {
		if another.Has(item) {
			ret.Add(item)
		}
		return true
	})
	return ret
}

// Diff .
func (s *Set) Diff(another *Set) *Set {
	ret := NewSet(s.Len())
	s.ForEach(func(item interface{}) bool {
		if !another.Has(item) {
			ret.Add(item)
		}
		return true
	})
	return ret
}

// Union .
func (s *Set) Union(another *Set) *Set {
	ret := NewSet(s.Len() + another.Len())
	s.ForEach(func(item interface{}) bool {
		ret.Add(item)
		return true
	})
	another.ForEach(func(item interface{}) bool {
		ret.Add(item)
		return true
	})
	return ret
}
