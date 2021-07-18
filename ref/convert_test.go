package ref

import (
	"reflect"
	"testing"
)

func TestStringCompare(t *testing.T) {
	cs := []struct {
		e bool
		a bool
		m string
	}{
		{true, interface{}("") != nil, `interface{}("") != nil`},
		{true, interface{}("") == "", `interface{}("") == ""`},
	}

	for _, c := range cs {
		if c.a != c.e {
			t.Error(c.m)
		}
	}
}

func TestConvert(t *testing.T) {
	cs := []struct {
		e interface{}
		s interface{}
		b reflect.Type
	}{
		{"1.123", 1.123, reflect.TypeOf("")},
		{int32(0777), "0777", reflect.TypeOf(int32(0))},
		{int(0), nil, reflect.TypeOf(int(0))},
		{int32(0), "", reflect.TypeOf(int32(0))},
	}

	for _, c := range cs {
		a, err := Convert(c.s, c.b)
		if err != nil {
			t.Errorf("Convert(%q, %q) Failed: %v, want %v", c.s, c.b, err, c.e)
		}
		if a != c.e {
			t.Errorf("Convert(%q, %q) = %v, want %v", c.s, c.b, a, c.e)
		}
	}
}
