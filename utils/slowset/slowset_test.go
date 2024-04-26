package slowset

import "testing"

func TestSet_Put(t *testing.T) {
	s := Make[int](10)
	s.Put(1)
	s.Put(1)
	s.Put(2)
	s.Put(2)
	s.Put(1)

	t.Log(s.elements)
}

func TestSet_Remove(t *testing.T) {
	s := Make[int](10)
	s.Put(1)
	s.Put(2)
	s.Put(3)

	t.Log(s.elements)

	s.Remove(1)

	t.Log(s.elements)
}

func TestSet_Remove2(t *testing.T) {
	s := Make[int](10)
	s.Put(1)
	s.Put(2)
	s.Put(3)

	t.Log(s.elements)

	s.Remove(2)

	t.Log(s.elements)
}

func TestSet_Remove3(t *testing.T) {
	s := Make[int](10)
	s.Put(1)
	s.Put(2)
	s.Put(3)

	t.Log(s.elements)

	s.Remove(3)

	t.Log(s.elements)
}
