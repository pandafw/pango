package col

import (
	"testing"
)

func TestOrderedMapItemString(t *testing.T) {
	cs := []struct {
		e string
		k interface{}
		v interface{}
	}{
		{"a => b", "a", "b"},
	}

	for i, c := range cs {
		a := (&OrderedMapItem{MapItem{c.k, c.v}, nil}).String()
		if a != c.e {
			t.Errorf("[%d] OrderedMapItem{MapItem{%v, %v}, nil}.String() = %q, want %q", i, c.k, c.v, a, c.e)
		}
	}
}
