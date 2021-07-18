package col

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/pandafw/pango/str"
)

type inta []interface{}

// Len is the number of elements in the collection.
func (a inta) Len() int {
	return len(a)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (a inta) Less(i, j int) bool {
	return a[i].(int) < a[j].(int)
}

// Swap swaps the elements with indexes i and j.
func (a inta) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func TestSortedListAsc(t *testing.T) {
	for i := 1; i < 100; i++ {
		sl := NewSortedList(func(a, b interface{}) bool {
			return a.(int) < b.(int)
		})
		a := make([]interface{}, 0, 100)
		for n := i; n < 100; n++ {
			a = append(a, n)
		}

		for j := 0; j < len(a); j++ {
			sl.Add(a[j])
			if !sort.IsSorted(inta(sl.Values())) {
				t.Errorf("[%d] sort.IsSorted(%v) = %v, want %v", j, sl.Values(), false, true)
			}
		}

		if !reflect.DeepEqual(a, sl.Values()) {
			t.Errorf("%v != %v", a, sl.Values())
		}
	}
}

func TestSortedListDesc(t *testing.T) {
	for i := 1; i < 100; i++ {
		sl := NewSortedList(LessInt)
		a := make([]interface{}, 0, 100)
		for n := i; n < 100; n++ {
			a = append(a, n)
		}
		for j := len(a) - 1; j >= 0; j-- {
			sl.Add(a[j])

			if !sort.IsSorted(inta(sl.Values())) {
				t.Errorf("[%d] sort.IsSorted(%v) = %v, want %v", j, sl.Values(), false, true)
			}
		}

		if !reflect.DeepEqual(a, sl.Values()) {
			t.Errorf("%v != %v", a, sl.Values())
		}
	}
}

func TestSortedListRandom(t *testing.T) {
	for i := 1; i < 100; i++ {
		sl := NewSortedList(func(a, b interface{}) bool {
			return a.(int) < b.(int)
		})
		a := make([]interface{}, 0, 100)
		for n := i; n < 100; n++ {
			a = append(a, rand.Intn(20))
		}
		for j := len(a) - 1; j >= 0; j-- {
			sl.Add(a[j])

			if !sort.IsSorted(inta(sl.Values())) {
				t.Errorf("[%d] sort.IsSorted(%v) = %v, want %v", j, sl.Values(), false, true)
			}
		}

		sort.Sort(inta(a))

		if !reflect.DeepEqual(a, sl.Values()) {
			t.Errorf("%v != %v", a, sl.Values())
		}
	}
}

func TestSortedListContains(t *testing.T) {
	sl := NewSortedList(LessString, "1", "11", "111", "1", "11", "111")

	n := str.Repeat("1", 3)

	if !sl.Contains(n) {
		t.Errorf("SortedList [%v] should contains %v", sl, n)
	}

	n += "1"
	if sl.Contains(n) {
		t.Errorf("SortedList [%v] should not contains %v", sl, n)
	}
}

func TestSetSearch(t *testing.T) {
	sl := NewSortedList(LessInt, 1, 11)

	n := 10 + 1
	sn, se := sl.Search(n)
	if se == nil || sn != 1 {
		t.Errorf("SortedList [%v] should contains %v", sl, n)
	}

	n = 1 + 10 + 100
	for i := 0; i < 100; i++ {
		n111 := sl.Add(111)
		sn, se = sl.Search(n)
		if se != n111 || sn != 2 {
			t.Errorf("SortedList [%v] should contains %v", sl, n)
		}
	}

	n++
	sn, se = sl.Search(n)
	if se != nil || sn != -1 {
		t.Errorf("SortedList [%v] should not contains %v", sl, n)
	}
}

func TestSortedListDelete(t *testing.T) {
	sl := NewSortedList(LessInt)

	for i := 0; i < 100; i++ {
		sl.Add(i)
	}

	if sl.Delete(100) {
		t.Error("sl.Delete(100)=true, want false")
	}
	for i := 0; i < 100; i++ {
		if !sl.Delete(i) {
			t.Errorf("sl.Delete(%v)=false, want true", i)
		}
	}
	if !sl.IsEmpty() {
		t.Error("sl.IsEmpty()=false, want true")
	}
}

func TestSortedListDeleteAll(t *testing.T) {
	sl := NewSortedList(LessInt)

	for i := 0; i < 100; i++ {
		z := i % 10
		for j := 0; j < z; j++ {
			sl.Add(i)
		}
	}

	if sl.DeleteAll(100) != 0 {
		t.Error("sl.DeleteAll(100) != 0")
	}
	for i := 0; i < 100; i++ {
		w := i % 10

		a := sl.DeleteAll(i)
		if a != w {
			t.Errorf("sl.DeleteAll(%d) = %v, want %v", i, a, w)
		}
	}
	if !sl.IsEmpty() {
		t.Error("sl.IsEmpty()=false, want true")
	}
}

func TestSortedListString(t *testing.T) {
	w := "[1,2,3]"
	a := fmt.Sprintf("%s", NewSortedList(LessInt, 1, 3, 2))
	if w != a {
		t.Errorf("SortedList.String() = %v, want %v", a, w)
	}
}

func TestSortedListMarshalJSON(t *testing.T) {
	type Case struct {
		list *SortedList
		json string
	}

	cs := []Case{
		{NewSortedList(LessString, "2", "1", "0"), `["0","1","2"]`},
	}

	for i, c := range cs {
		bs, err := json.Marshal(c.list)

		if err != nil {
			t.Fatalf("[%d] json.Marshal(%v) error: %v", i, c.list, err)
		}

		a := string(bs)
		if !reflect.DeepEqual(c.json, a) {
			t.Fatalf("[%d] json.Marshal(%v) = %v, want %v", i, c.list, a, c.json)
		}
	}
}

func TestSortedListUnmarshalJSON(t *testing.T) {
	cs := []struct {
		json string
		list *List
	}{
		{`["2","0","1"]`, NewList("0", "1", "2")},
	}

	for i, c := range cs {
		a := NewSortedList(LessString)
		err := json.Unmarshal([]byte(c.json), a)

		if err != nil {
			t.Fatalf("[%d] json.Unmarshal(%v) error: %v", i, c.json, err)
		}

		if !reflect.DeepEqual(a.list, c.list) {
			t.Fatalf("[%d] json.Marshal(%v) = %v, want %v", i, c.json, a.list, c.list)
		}
	}
}
