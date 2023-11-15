package directstring

import "testing"

func TestCastFromString(t *testing.T) {
	var s = "hello"
	ds := CastFromString(s)
	t.Log(s, ds.ToString())
}

func TestHashCode(t *testing.T) {
	var s = "hello"
	ds := CastFromString(s)
	t.Log(ds.HashCode())

	var s2 = "hello, world"
	ds2 := CastFromString(s2)
	t.Log(ds2.HashCode())
}
