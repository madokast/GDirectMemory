package slowset

import "fmt"

type Set[E comparable] struct {
	elements []E
}

func Make[E comparable](size int) Set[E] {
	return Set[E]{
		elements: make([]E, 0, size),
	}
}

func (s *Set[E]) Put(e E) {
	for _, e2 := range s.elements {
		if e == e2 {
			return
		}
	}
	if cap(s.elements) == len(s.elements) {
		panic("set full")
	}
	s.elements = append(s.elements, e)
}

func (s *Set[E]) DistinctPut(e E) {
	for _, e2 := range s.elements {
		if e == e2 {
			panic(fmt.Sprintf("duplicate %v %s", e, s.String()))
		}
	}
	if cap(s.elements) == len(s.elements) {
		panic("set full")
	}
	s.elements = append(s.elements, e)
}

func (s *Set[E]) MustRemove(e E) {
	for i, e2 := range s.elements {
		if e == e2 {
			copy(s.elements[i:], s.elements[i+1:])
			s.elements = s.elements[:len(s.elements)-1]
			return
		}
	}
	panic(fmt.Sprintf("no %v in %s", e, s.String()))
}

func (s *Set[E]) Remove(e E) {
	for i, e2 := range s.elements {
		if e == e2 {
			copy(s.elements[i:], s.elements[i+1:])
			s.elements = s.elements[:len(s.elements)-1]
			return
		}
	}
}

func (s *Set[E]) String() string {
	return fmt.Sprintf("%v", s.elements)
}
