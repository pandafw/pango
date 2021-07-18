package col

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/pandafw/pango/str"
)

func checkListLen(t *testing.T, l *List, len int) bool {
	if n := l.Len(); n != len {
		t.Errorf("l.Len() = %d, want %d", n, len)
		return false
	}
	return true
}

func checkListPointers(t *testing.T, l *List, lis []*ListItem) {
	root := &l.root

	if !checkListLen(t, l, len(lis)) {
		return
	}

	// zero length lists must be the zero value or properly initialized (sentinel circle)
	if len(lis) == 0 {
		if l.root.next != nil && l.root.next != root || l.root.prev != nil && l.root.prev != root {
			t.Errorf("l.root.next = %p, l.root.prev = %p; both should both be nil or %p", l.root.next, l.root.prev, root)
		}
		return
	}
	// len(lis) > 0

	// check internal and external prev/next connections
	for i, e := range lis {
		prev := root
		Prev := (*ListItem)(nil)
		if i > 0 {
			prev = lis[i-1]
			Prev = prev
		}
		if p := e.prev; p != prev {
			t.Errorf("elt[%d](%p).prev = %p, want %p", i, e, p, prev)
		}
		if p := e.Prev(); p != Prev {
			t.Errorf("elt[%d](%p).Prev() = %p, want %p", i, e, p, Prev)
		}

		next := root
		Next := (*ListItem)(nil)
		if i < len(lis)-1 {
			next = lis[i+1]
			Next = next
		}
		if n := e.next; n != next {
			t.Errorf("elt[%d](%p).next = %p, want %p", i, e, n, next)
		}
		if n := e.Next(); n != Next {
			t.Errorf("elt[%d](%p).Next() = %p, want %p", i, e, n, Next)
		}
	}
}

func TestList(t *testing.T) {
	l := NewList()
	checkListPointers(t, l, []*ListItem{})

	// Single item list
	e := l.PushFront("a")
	checkListPointers(t, l, []*ListItem{e})
	l.MoveToFront(e)
	checkListPointers(t, l, []*ListItem{e})
	l.MoveToBack(e)
	checkListPointers(t, l, []*ListItem{e})
	l.Remove(e)
	checkListPointers(t, l, []*ListItem{})

	// Bigger list
	e2 := l.PushFront(2)
	e1 := l.PushFront(1)
	e3 := l.PushBack(3)
	e4 := l.PushBack("banana")
	checkListPointers(t, l, []*ListItem{e1, e2, e3, e4})

	l.Remove(e2)
	checkListPointers(t, l, []*ListItem{e1, e3, e4})

	l.MoveToFront(e3) // move from middle
	checkListPointers(t, l, []*ListItem{e3, e1, e4})

	l.MoveToFront(e1)
	l.MoveToBack(e3) // move from middle
	checkListPointers(t, l, []*ListItem{e1, e4, e3})

	l.MoveToFront(e3) // move from back
	checkListPointers(t, l, []*ListItem{e3, e1, e4})
	l.MoveToFront(e3) // should be no-op
	checkListPointers(t, l, []*ListItem{e3, e1, e4})

	l.MoveToBack(e3) // move from front
	checkListPointers(t, l, []*ListItem{e1, e4, e3})
	l.MoveToBack(e3) // should be no-op
	checkListPointers(t, l, []*ListItem{e1, e4, e3})

	e2 = l.InsertBefore(2, e1) // insert before front
	checkListPointers(t, l, []*ListItem{e2, e1, e4, e3})
	l.Remove(e2)
	e2 = l.InsertBefore(2, e4) // insert before middle
	checkListPointers(t, l, []*ListItem{e1, e2, e4, e3})
	l.Remove(e2)
	e2 = l.InsertBefore(2, e3) // insert before back
	checkListPointers(t, l, []*ListItem{e1, e4, e2, e3})
	l.Remove(e2)

	e2 = l.InsertAfter(2, e1) // insert after front
	checkListPointers(t, l, []*ListItem{e1, e2, e4, e3})
	l.Remove(e2)
	e2 = l.InsertAfter(2, e4) // insert after middle
	checkListPointers(t, l, []*ListItem{e1, e4, e2, e3})
	l.Remove(e2)
	e2 = l.InsertAfter(2, e3) // insert after back
	checkListPointers(t, l, []*ListItem{e1, e4, e3, e2})
	l.Remove(e2)

	// Check standard iteration.
	sum := 0
	for e := l.Front(); e != nil; e = e.Next() {
		if i, ok := e.Value.(int); ok {
			sum += i
		}
	}
	if sum != 4 {
		t.Errorf("sum over l = %d, want 4", sum)
	}

	// Clear all items by iterating
	var next *ListItem
	for e := l.Front(); e != nil; e = next {
		next = e.Next()
		l.Remove(e)
	}
	checkListPointers(t, l, []*ListItem{})
}

func checkList(t *testing.T, l *List, evs []interface{}) {
	if !checkListLen(t, l, len(evs)) {
		return
	}

	for i, e := 0, l.Front(); e != nil; i, e = i+1, e.Next() {
		v := e.Value.(int)
		if v != evs[i] {
			t.Errorf("elt[%d].Value = %v, want %v", i, v, evs[i])
		}
	}

	avs := l.Values()
	for i, v := range avs {
		if v != evs[i] {
			t.Errorf("elt[%d].Value = %v, want %v", i, v, evs[i])
		}
	}
}

func TestListExtending(t *testing.T) {
	l1 := NewList(1, 2, 3)
	l2 := NewList()
	l2.PushBack(4)
	l2.PushBack(5)

	l3 := NewList()
	l3.PushBackList(l1)
	checkList(t, l3, []interface{}{1, 2, 3})
	l3.PushBackList(l2)
	checkList(t, l3, []interface{}{1, 2, 3, 4, 5})

	l3 = NewList()
	l3.PushFrontList(l2)
	checkList(t, l3, []interface{}{4, 5})
	l3.PushFrontList(l1)
	checkList(t, l3, []interface{}{1, 2, 3, 4, 5})

	checkList(t, l1, []interface{}{1, 2, 3})
	checkList(t, l2, []interface{}{4, 5})

	l3 = NewList()
	l3.PushBackList(l1)
	checkList(t, l3, []interface{}{1, 2, 3})
	l3.PushBackList(l3)
	checkList(t, l3, []interface{}{1, 2, 3, 1, 2, 3})

	l3 = NewList()
	l3.PushFrontList(l1)
	checkList(t, l3, []interface{}{1, 2, 3})
	l3.PushFrontList(l3)
	checkList(t, l3, []interface{}{1, 2, 3, 1, 2, 3})

	l3 = NewList()
	l1.PushBackList(l3)
	checkList(t, l1, []interface{}{1, 2, 3})
	l1.PushFrontList(l3)
	checkList(t, l1, []interface{}{1, 2, 3})

	l1.Clear()
	l2.Clear()
	l3.Clear()
	l1.PushBackAll(1, 2, 3)
	checkList(t, l1, []interface{}{1, 2, 3})
	l2.PushBackAll(4, 5)
	checkList(t, l2, []interface{}{4, 5})
	l3.PushBackList(l1)
	checkList(t, l3, []interface{}{1, 2, 3})
	l3.PushBackAll(4, 5)
	checkList(t, l3, []interface{}{1, 2, 3, 4, 5})
	l3.PushFrontAll(4, 5)
	checkList(t, l3, []interface{}{4, 5, 1, 2, 3, 4, 5})
}

func TestListRemove(t *testing.T) {
	l := NewList()
	e1 := l.PushBack(1)
	e2 := l.PushBack(2)
	checkListPointers(t, l, []*ListItem{e1, e2})
	e := l.Front()
	l.Remove(e)
	checkListPointers(t, l, []*ListItem{e2})
	l.Remove(e)
	checkListPointers(t, l, []*ListItem{e2})
}

func TestListInsertBefore(t *testing.T) {
	l1 := NewList()
	l1.PushBack(1)
	l1.PushBack(2)

	l2 := NewList()
	l2.PushBack(3)
	l2.PushBack(4)

	e := l1.Front()
	l2.Remove(e) // l2 should not change because e is not an item of l2
	if n := l2.Len(); n != 2 {
		t.Errorf("l2.Len() = %d, want 2", n)
	}

	l1.InsertBefore(8, e)
	if n := l1.Len(); n != 3 {
		t.Errorf("l1.Len() = %d, want 3", n)
	}
}

func TestListRemove1(t *testing.T) {
	l := NewList()
	l.PushBack(1)
	l.PushBack(2)

	e := l.Front()
	l.Remove(e)
	if e.Value != 1 {
		t.Errorf("e.value = %d, want 1", e.Value)
	}
	if e.Next() != nil {
		t.Errorf("e.Next() != nil")
	}
	if e.Prev() != nil {
		t.Errorf("e.Prev() != nil")
	}
}

func TestListMove(t *testing.T) {
	l := NewList()
	e1 := l.PushBack(1)
	e2 := l.PushBack(2)
	e3 := l.PushBack(3)
	e4 := l.PushBack(4)

	l.MoveAfter(e3, e3)
	checkListPointers(t, l, []*ListItem{e1, e2, e3, e4})
	l.MoveBefore(e2, e2)
	checkListPointers(t, l, []*ListItem{e1, e2, e3, e4})

	l.MoveAfter(e3, e2)
	checkListPointers(t, l, []*ListItem{e1, e2, e3, e4})
	l.MoveBefore(e2, e3)
	checkListPointers(t, l, []*ListItem{e1, e2, e3, e4})

	l.MoveBefore(e2, e4)
	checkListPointers(t, l, []*ListItem{e1, e3, e2, e4})
	e2, e3 = e3, e2

	l.MoveBefore(e4, e1)
	checkListPointers(t, l, []*ListItem{e4, e1, e2, e3})
	e1, e2, e3, e4 = e4, e1, e2, e3

	l.MoveAfter(e4, e1)
	checkListPointers(t, l, []*ListItem{e1, e4, e2, e3})
	e2, e3, e4 = e4, e2, e3

	l.MoveAfter(e2, e3)
	checkListPointers(t, l, []*ListItem{e1, e3, e2, e4})
	e2, e3 = e3, e2
}

// Test that a list l is not modified when calling InsertBefore with a mark that is not an item of l.
func TestListInsertBeforeUnknownMark(t *testing.T) {
	l := NewList()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.InsertBefore(1, new(ListItem))
	checkList(t, l, []interface{}{1, 2, 3})
}

// Test that a list l is not modified when calling InsertAfter with a mark that is not an item of l.
func TestListInsertAfterUnknownMark(t *testing.T) {
	l := NewList()
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	l.InsertAfter(1, new(ListItem))
	checkList(t, l, []interface{}{1, 2, 3})
}

// Test that a list l is not modified when calling MoveAfter or MoveBefore with a mark that is not an item of l.
func TestListMoveUnknownMark(t *testing.T) {
	l1 := NewList()
	e1 := l1.PushBack(1)

	l2 := NewList()
	e2 := l2.PushBack(2)

	l1.MoveAfter(e1, e2)
	checkList(t, l1, []interface{}{1})
	checkList(t, l2, []interface{}{2})

	l1.MoveBefore(e1, e2)
	checkList(t, l1, []interface{}{1})
	checkList(t, l2, []interface{}{2})
}

func TestListContains(t *testing.T) {
	l := NewList(1, 11, 111, "1", "11", "111")

	n := (100+1)/101 + 110

	if !l.Contains(n) {
		t.Errorf("List [%v] should contains %v", l, n)
	}

	n++
	if l.Contains(n) {
		t.Errorf("List [%v] should not contains %v", l, n)
	}

	s := str.Repeat("1", 3)

	if !l.Contains(s) {
		t.Errorf("List [%v] should contains %v", l, s)
	}

	s += "0"
	if l.Contains(s) {
		t.Errorf("List [%v] should not contains %v", l, s)
	}
}

func TestListItem(t *testing.T) {
	l := NewList()

	for i := 0; i < 100; i++ {
		l.PushBack(i)
	}

	if l.Item(100) != nil {
		t.Error("List.Item(100) != nil")
	}
	if l.Item(-101) != nil {
		t.Error("List.Item(-101) != nil")
	}

	for i := 0; i < 100; i++ {
		w := i
		s := i
		a := l.Item(s).Value
		if a != w {
			t.Errorf("List.Item(%v).Value = %v, want %v", s, a, w)
		}

		s = i - 100
		a = l.Item(s).Value
		if a != w {
			t.Errorf("List.Item(%v).Value = %v, want %v", s, a, w)
		}
	}
}

func TestListSwap(t *testing.T) {
	l := NewList()

	for i := 0; i < 100; i++ {
		l.PushBack(i)
	}

	if l.Swap(l.Item(0), l.Front()) {
		t.Error("l.Swap(l.Item(0), l.Front()) = true, want false")
	}

	for i := 0; i < 50; i++ {
		ia := l.Item(i)
		ib := l.Item(-i - 1)
		a := l.Swap(ia, ib)
		if !a {
			t.Errorf("List.Swap(%v, %v) = %v, want %v", ia, ib, a, true)
		}
	}

	for i := 0; i < 100; i++ {
		e := 100 - 1 - i
		a := l.Item(i).Value
		if a != e {
			t.Errorf("List.Item(%v).Value = %v, want %v", i, a, e)
		}
	}
}

func TestListSearch(t *testing.T) {
	l := NewList(1, 11)

	n111 := l.PushBack(111)

	l.PushBackAll("1", "11")
	s111 := l.PushBack("111")

	n := (100+1)/101 + 110

	sn, se := l.Search(n)
	if n111 != se || sn != 2 {
		t.Errorf("List [%v] should contains %v", l, n)
	}

	n++
	sn, se = l.Search(n)
	if se != nil || sn != -1 {
		t.Errorf("List [%v] should not contains %v", l, n)
	}

	s := str.Repeat("1", 3)

	sn, se = l.Search(s)
	if s111 != se || sn != 5 {
		t.Errorf("List [%v] should contains %v", l, s)
	}

	s += "0"
	sn, se = l.Search(s)
	if se != nil || sn != -1 {
		t.Errorf("List [%v] should not contains %v", l, s)
	}
}

func TestListDelete(t *testing.T) {
	l := NewList()

	for i := 0; i < 100; i++ {
		l.PushBack(i)
	}

	if l.Delete(100) {
		t.Error("List.Delete(100) should return false")
	}
	for i := 0; i < 100; i++ {
		if !l.Delete(i) {
			t.Errorf("List.Delete(%v) should return true", i)
		}
	}

	if !l.IsEmpty() {
		t.Error("List.Empty() should return true")
	}
}

func TestListDeleteAll(t *testing.T) {
	l := NewList()

	for i := 0; i < 100; i++ {
		z := i % 10
		for j := 0; j < z; j++ {
			l.PushBack(i)
		}
	}

	if l.DeleteAll(100) != 0 {
		t.Error("List.DeleteAll(100) != 0")
	}
	for i := 0; i < 100; i++ {
		z := i % 10
		a := l.DeleteAll(i)
		if a != z {
			t.Errorf("List.DeleteAll(%v) = %v, want %v", i, a, z)
		}
	}

	if !l.IsEmpty() {
		t.Error("List.Empty() should return true")
	}
}

func TestListString(t *testing.T) {
	e := "[1,3,2]"
	a := fmt.Sprintf("%s", NewList(1, 3, 2))
	if a != e {
		t.Errorf(`fmt.Sprintf("%%s", NewList(1, 3, 2)) = %v, want %v`, a, e)
	}
}

func TestListMarshalJSON(t *testing.T) {
	cs := []struct {
		list *List
		json string
	}{
		{NewList(0, 1, "0", "1", 0.0, 1.0, true, false), `[0,1,"0","1",0,1,true,false]`},
		{NewList(0, "1", 2.0, []int{1, 2}, map[int]int{1: 10, 2: 20}), `[0,"1",2,[1,2],{"1":10,"2":20}]`},
	}

	for i, c := range cs {
		bs, err := json.Marshal(c.list)
		if err != nil {
			t.Errorf("[%d] json.Marshal(%v) error: %v", i, c.list, err)
		}

		a := string(bs)
		if a != c.json {
			t.Errorf("[%d] json.Marshal(%v) = %q, want %q", i, c.list, a, c.list)
		}
	}
}

func TestListUnmarshalJSON(t *testing.T) {
	type Case struct {
		json string
		list *List
	}

	cs := []Case{
		{`["0","1",0,1,true,false]`, NewList("0", "1", 0.0, 1.0, true, false)},
		{`["1",2,[1,2],{"1":10,"2":20}]`, NewList("1", 2.0, NewList(1.0, 2.0), JSONObject{"1": 10.0, "2": 20.0})},
	}

	for i, c := range cs {
		a := NewList()
		err := json.Unmarshal([]byte(c.json), a)

		if err != nil {
			t.Errorf("[%d] json.Unmarshal(%v) error: %v", i, c.json, err)
		}

		if !reflect.DeepEqual(a, c.list) {
			t.Errorf("[%d] json.Unmarshal(%q) = %v, want %q", i, c.json, a, c.list)
		}
	}
}
