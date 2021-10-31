package flock

type Set struct {
	list map[interface{}]struct{} //empty structs occupy 0 memory
}

func (s *Set) Has(v interface{}) bool {
	_, ok := s.list[v]
	return ok
}

func (s *Set) Add(v interface{}) {
	s.list[v] = struct{}{}
}

func (s *Set) Remove(v interface{}) {
	delete(s.list, v)
}

func (s *Set) Clear() {
	s.list = make(map[interface{}]struct{})
}

func (s *Set) Size() interface{} {
	return len(s.list)
}

func NewSet() *Set {
	s := &Set{}
	s.list = make(map[interface{}]struct{})
	return s
}

//optional functionalities

//AddMulti Add multiple values in the set
func (s *Set) AddMulti(list ...interface{}) {
	for _, v := range list {
		s.Add(v)
	}
}

type FilterFunc func(v interface{}) bool

// Filter returns a subset, that contains only the values that satisfies the given predicate P
func (s *Set) Filter(P FilterFunc) *Set {
	res := NewSet()
	for v := range s.list {
		if P(v) == false {
			continue
		}
		res.Add(v)
	}
	return res
}

func (s *Set) Union(s2 *Set) *Set {
	res := NewSet()
	for v := range s.list {
		res.Add(v)
	}

	for v := range s2.list {
		res.Add(v)
	}
	return res
}

func (s *Set) Intersect(s2 *Set) *Set {
	res := NewSet()
	for v := range s.list {
		if s2.Has(v) == false {
			continue
		}
		res.Add(v)
	}
	return res
}

// Difference returns the subset from s, that doesn't exists in s2 (param)
func (s *Set) Difference(s2 *Set) *Set {
	res := NewSet()
	for v := range s.list {
		if s2.Has(v) {
			continue
		}
		res.Add(v)
	}
	return res
}
