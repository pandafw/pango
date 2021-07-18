package col

import (
	"testing"
)

func TestMapItemString(t *testing.T) {
	cs := []struct {
		e string
		k interface{}
		v interface{}
	}{
		{"a => b", "a", "b"},
	}

	for i, c := range cs {
		a := (&MapItem{c.k, c.v}).String()
		if a != c.e {
			t.Errorf("[%d] MapItem{%v, %v}.String() = %q, want %q", i, c.k, c.v, a, c.e)
		}
	}
}
