package stdlib

import (
	"testing"
)

func TestMalloc(t *testing.T) {
	m := Malloc(128)
	t.Log("Obtain mempry", m.ToString())
	m.Free()
}
