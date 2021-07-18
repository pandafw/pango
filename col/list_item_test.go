package col

import (
	"testing"
)

func TestListItemString(t *testing.T) {
	cs := []struct {
		e string
		s interface{}
	}{
		{"a", "a"},
		{"1", 1},
	}

	for _, c := range cs {
		a := (&ListItem{Value: c.s}).String()
		if a != c.e {
			t.Errorf("ListIem(%v).String() = %q, want %q", c.s, a, c.e)
		}
	}
}
