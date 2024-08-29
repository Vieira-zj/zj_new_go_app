package structs

type Set[E comparable] struct {
	m map[E]struct{}
}

func NewSet[E comparable]() *Set[E] {
	return &Set[E]{
		m: make(map[E]struct{}),
	}
}

func (s Set[E]) Add(v E) {
	s.m[v] = struct{}{}
}

func (s Set[E]) Contains(v E) bool {
	_, ok := s.m[v]
	return ok
}

func (Set[E]) Union(s1, s2 *Set[E]) *Set[E] {
	ret := NewSet[E]()
	for v := range s1.m {
		ret.Add(v)
	}
	for v := range s2.m {
		ret.Add(v)
	}

	return ret
}

func (s Set[E]) Range(fn func(v E) bool) {
	for v := range s.m {
		if !fn(v) {
			return
		}
	}
}

func (s Set[E]) Iterator() (func() (E, bool), func()) {
	ch := make(chan E)
	stopCh := make(chan struct{})

	go func() {
		defer close(ch)
		for v := range s.m {
			select {
			case ch <- v:
			case <-stopCh:
				return
			}
		}
	}()

	next := func() (E, bool) {
		v, ok := <-ch
		return v, ok
	}
	stop := func() {
		close(stopCh)
	}

	return next, stop
}
