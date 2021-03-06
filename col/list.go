package col

import (
	"encoding/json"
)

// List implements a doubly linked list.
// The zero value for List is an empty list ready to use.
//
// To iterate over a list (where l is a *List):
//	for li := l.Front(); li != nil; li = li.Next() {
//		// do something with li.Value
//	}
//
type List struct {
	root ListItem // sentinel list item, only &root, root.prev, and root.next are used
	len  int      // current list length excluding (this) sentinel item
}

// NewList returns an initialized list.
// Example: NewList(1, 2, 3)
func NewList(vs ...interface{}) *List {
	l := &List{}
	l.Clear()
	l.PushBackAll(vs...)
	return l
}

// NewStrList returns an initialized list.
// Example: NewList("1", "2", "3")
func NewStrList(ss ...string) *List {
	l := &List{}
	l.Clear()
	for _, s := range ss {
		l.insertValue(s, l.root.prev)
	}
	return l
}

// Len returns the length of the list.
// The complexity is O(1).
func (l *List) Len() int {
	return l.len
}

// IsEmpty returns true if the list length == 0
func (l *List) IsEmpty() bool {
	return l.len == 0
}

// Item returns the item at the specified index
// if i < -l.Len() or i >= l.Len(), returns nil
// if i < 0, returns l.Item(l.Len() + i)
func (l *List) Item(i int) *ListItem {
	if i < -l.len || i >= l.len {
		return nil
	}

	if i < 0 {
		i += l.len
	}
	if i >= l.len/2 {
		return l.Back().Offset(i + 1 - l.len)
	}

	return l.Front().Offset(i)
}

// Front returns the first item of list l or nil if the list is empty.
func (l *List) Front() *ListItem {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last item of list l or nil if the list is empty.
func (l *List) Back() *ListItem {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// Contains Test to see whether or not the v is in the list
func (l *List) Contains(v interface{}) bool {
	_, li := l.Search(v)
	return li != nil
}

// Search linear search v
// returns index, item if it's value is v
// if not found, returns -1, nil
func (l *List) Search(v interface{}) (int, *ListItem) {
	for i, li := 0, l.Front(); li != nil; li = li.Next() {
		if li.Value == v {
			return i, li
		}
		i++
	}
	return -1, nil
}

// Clear clears list l.
func (l *List) Clear() {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
}

// insert inserts item li after at, increments l.len, and returns li.
func (l *List) insert(li, at *ListItem) *ListItem {
	ni := at.next
	at.next = li
	li.prev = at
	li.next = ni
	ni.prev = li
	li.list = l
	l.len++
	return li
}

// insertValue is a convenience wrapper for insert(&ListItem{Value: v}, at).
func (l *List) insertValue(v interface{}, at *ListItem) *ListItem {
	return l.insert(&ListItem{Value: v}, at)
}

// remove removes the item li from its list, decrements l.len, and returns li.
func (l *List) remove(li *ListItem) *ListItem {
	li.prev.next = li.next
	li.next.prev = li.prev
	li.next = nil // avoid memory leaks
	li.prev = nil // avoid memory leaks
	li.list = nil
	l.len--
	return li
}

// move moves the item li to next to at and returns li.
func (l *List) move(li, at *ListItem) *ListItem {
	if li == at {
		return li
	}
	li.prev.next = li.next
	li.next.prev = li.prev

	n := at.next
	at.next = li
	li.prev = at
	li.next = n
	n.prev = li

	return li
}

// Delete delete the first item with associated value v
// returns true if v is in the list
// returns false if the the list is not changed
func (l *List) Delete(v interface{}) bool {
	_, li := l.Search(v)
	if li != nil {
		l.remove(li)
		return true
	}

	return false
}

// DeleteAll delete all items with associated value v
// returns the deleted count
func (l *List) DeleteAll(v interface{}) int {
	n := 0
	for li := l.Front(); li != nil; {
		ni := li.Next()
		if li.Value == v {
			l.remove(li)
			n++
		}
		li = ni
	}

	return n
}

// Remove removes the item li from l if li is an item of list l.
// The item li must not be nil.
func (l *List) Remove(li *ListItem) {
	if li.list == l {
		// if li.list == l, l must have been initialized when li was inserted
		// in l or l == nil (li is a zero item) and l.remove will crash
		l.remove(li)
	}
}

// PushFront inserts a new item li with value v at the front of list l and returns li.
func (l *List) PushFront(v interface{}) *ListItem {
	return l.insertValue(v, &l.root)
}

// PushFrontAll inserts all items of vs at the front of list l.
func (l *List) PushFrontAll(vs ...interface{}) {
	li := &l.root
	for _, v := range vs {
		li = l.insertValue(v, li)
	}
}

// PushFrontList inserts a copy of an other list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List) PushFrontList(other *List) {
	for i, li := other.Len(), other.Back(); i > 0; i, li = i-1, li.Prev() {
		l.insertValue(li.Value, &l.root)
	}
}

// PushBack inserts a new item li with value v at the back of list l and returns li.
func (l *List) PushBack(v interface{}) *ListItem {
	return l.insertValue(v, l.root.prev)
}

// PushBackAll inserts all items of vs at the back of list l.
func (l *List) PushBackAll(vs ...interface{}) {
	for _, v := range vs {
		l.insertValue(v, l.root.prev)
	}
}

// PushBackList inserts a copy of an other list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List) PushBackList(other *List) {
	for i, li := other.Len(), other.Front(); i > 0; i, li = i-1, li.Next() {
		l.insertValue(li.Value, l.root.prev)
	}
}

// InsertBefore inserts a new item li with value v immediately before at and returns li.
// If at is not an item of l, the list is not modified.
// The at must not be nil.
func (l *List) InsertBefore(v interface{}, at *ListItem) *ListItem {
	if at.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, at.prev)
}

// InsertAfter inserts a new item li with value v immediately after at and returns li.
// If at is not an item of l, the list is not modified.
// The at must not be nil.
func (l *List) InsertAfter(v interface{}, at *ListItem) *ListItem {
	if at.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, at)
}

// MoveToFront moves item li to the front of list l.
// If li is not an item of l, the list is not modified.
// The item must not be nil.
// Returns true if list is modified.
func (l *List) MoveToFront(li *ListItem) bool {
	if li.list != l || l.root.next == li {
		return false
	}
	// see comment in List.Remove about initialization of l
	l.move(li, &l.root)
	return true
}

// MoveToBack moves item li to the back of list l.
// If li is not an item of l, the list is not modified.
// The item must not be nil.
// Returns true if list is modified.
func (l *List) MoveToBack(li *ListItem) bool {
	if li.list != l || l.root.prev == li {
		return false
	}
	// see comment in List.Remove about initialization of l
	l.move(li, l.root.prev)
	return true
}

// MoveBefore moves item li to its new position before at.
// If li or at is not an item of l, or li == at, the list is not modified.
// The item and at must not be nil.
// Returns true if list is modified.
func (l *List) MoveBefore(li, at *ListItem) bool {
	if li.list != l || li == at || at.list != l {
		return false
	}
	l.move(li, at.prev)
	return true
}

// MoveAfter moves item li to its new position after at.
// If li or at is not an item of l, or li == at, the list is not modified.
// The item and at must not be nil.
// Returns true if list is modified.
func (l *List) MoveAfter(li, at *ListItem) bool {
	if li.list != l || li == at || at.list != l {
		return false
	}
	l.move(li, at)
	return true
}

// Swap swap item's value of ia, ib.
// If ia or ib is not an item of l, or ia == ib, the list is not modified.
// The item and at must not be nil.
// Returns true if list is modified.
func (l *List) Swap(ia, ib *ListItem) bool {
	if ia.list != l || ia == ib || ib.list != l {
		return false
	}
	ia.Value, ib.Value = ib.Value, ia.Value
	return true
}

// Values returns a slice contains all the items of the list l
func (l *List) Values() []interface{} {
	a := make([]interface{}, 0, l.Len())
	for li := l.Front(); li != nil; li = li.Next() {
		a = append(a, li.Value)
	}
	return a
}

// Each Call f for each item in the set
func (l *List) Each(f func(interface{})) {
	for li := l.Front(); li != nil; li = li.Next() {
		f(li.Value)
	}
}

// ReverseEach Call f for each item in the set with reverse order
func (l *List) ReverseEach(f func(interface{})) {
	for li := l.Back(); li != nil; li = li.Prev() {
		f(li.Value)
	}
}

// String print list to string
func (l *List) String() string {
	bs, _ := json.Marshal(l)
	return string(bs)
}

/*------------- JSON -----------------*/

func newJSONArrayList() jsonArray {
	return NewList()
}

func (l *List) addJSONArrayItem(v interface{}) jsonArray {
	l.PushBack(v)
	return l
}

// MarshalJSON implements type json.Marshaler interface, so can be called in json.Marshal(l)
func (l *List) MarshalJSON() (res []byte, err error) {
	if l.IsEmpty() {
		return []byte("[]"), nil
	}

	res = append(res, '[')
	for li := l.Front(); li != nil; li = li.Next() {
		var b []byte
		b, err = json.Marshal(li.Value)
		if err != nil {
			return
		}
		res = append(res, b...)
		res = append(res, ',')
	}
	res[len(res)-1] = ']'
	return
}

// UnmarshalJSON implements type json.Unmarshaler interface, so can be called in json.Unmarshal(data, l)
func (l *List) UnmarshalJSON(data []byte) error {
	ju := &jsonUnmarshaler{
		newArray:  newJSONArrayList,
		newObject: newJSONObject,
	}
	return ju.unmarshalJSONArray(data, l)
}
