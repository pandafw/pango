package col

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

func TestOrderedMapBasicFeatures(t *testing.T) {
	n := 100
	om := NewOrderedMap()

	// set(i, 2 * i)
	for i := 0; i < n; i++ {
		ov, ok := interface{}(nil), false

		assertLenEqual("TestOrderedMapBasicFeatures", t, om, i)
		if i%2 == 0 {
			ov, ok = om.Set(i, 2*i)
		} else {
			ov, ok = om.SetIfAbsent(i, 2*i)
		}
		assertLenEqual("TestOrderedMapBasicFeatures", t, om, i+1)

		var w interface{}
		if ov != w {
			t.Errorf("[%d] set val = %v, want %v", i, ov, w)
		}
		w = false
		if ok != w {
			t.Errorf("[%d] set ok = %v, want %v", i, ok, w)
		}

		ov, ok = om.SetIfAbsent(i, 3*i)
		w = 2 * i
		if ov != w {
			t.Errorf("[%d] set val = %v, want %v", i, ov, w)
		}
		w = true
		if ok != w {
			t.Errorf("[%d] set ok = %v, want %v", i, ok, w)
		}
	}

	// get what we just set
	for i := 0; i < n; i++ {
		ov, ok := om.Get(i)

		var w interface{}
		w = 2 * i
		if ov != w {
			t.Errorf("[%d] get val = %v, want %v", i, ov, w)
		}
		w = true
		if ok != w {
			t.Errorf("[%d] get ok = %v, want %v", i, ok, w)
		}
	}

	// get items of what we just set
	for i := 0; i < n; i++ {
		item := om.Item(i)

		if item == nil {
			t.Errorf("[%d] item = %v, want %v", i, item, "not nil")
		}
		w := 2 * i
		if item.Value != w {
			t.Errorf("[%d] item.Value = %v, want %v", i, item.Value, w)
		}
	}

	// keys
	ks := make([]interface{}, n)
	for i := 0; i < n; i++ {
		ks[i] = i
	}
	if !reflect.DeepEqual(ks, om.Keys()) {
		t.Errorf("om.Keys() = %v, want %v", om.Keys(), ks)
	}

	// items
	mis := om.Items()
	if n != len(mis) {
		t.Errorf("len(mis) = %v, want %v", len(mis), n)
	}
	for i := 0; i < n; i++ {
		if i != mis[i].Key() {
			t.Errorf("mis[%d].Key() = %v, want %v", i, mis[i].Key(), i)
		}
		if i*2 != mis[i].Value {
			t.Errorf("mis[%d].Value = %v, want %v", i, mis[i].Value, i*2)
		}
	}

	// values
	vs := make([]interface{}, n)
	for i := 0; i < n; i++ {
		vs[i] = i * 2
	}
	if !reflect.DeepEqual(vs, om.Values()) {
		t.Errorf("om.Values() = %v, want %v", om.Values(), vs)
	}

	// forward iteration
	i := 0
	for mi := om.Front(); mi != nil; mi = mi.Next() {
		if i != mi.key {
			t.Errorf("[%d] mi.Key = %v, want %v", i, mi.key, i)
		}
		if i*2 != mi.Value {
			t.Errorf("[%d] mi.Value = %v, want %v", i, mi.Value, i*2)
		}
		i++
	}

	// backward iteration
	i = n - 1
	for mi := om.Back(); mi != nil; mi = mi.Prev() {
		if i != mi.key {
			t.Errorf("[%d] mi.Key = %v, want %v", i, mi.key, i)
		}
		if i*2 != mi.Value {
			t.Errorf("[%d] mi.Value = %v, want %v", i, mi.Value, i*2)
		}
		i--
	}

	// forward iteration starting from known key
	i = 42
	for mi := om.Item(i); mi != nil; mi = mi.Next() {
		if i != mi.key {
			t.Errorf("[%d] mi.Key = %v, want %v", i, mi.key, i)
		}
		if i*2 != mi.Value {
			t.Errorf("[%d] mi.Value = %v, want %v", i, mi.Value, i*2)
		}
		i++
	}

	// double values for items with even keys
	for j := 0; j < n/2; j++ {
		i = 2 * j
		ov, ok := om.Set(i, 4*i)

		if 2*i != ov {
			t.Errorf("[%d] set val = %v, want %v", i, ov, 2*i)
		}
		if !ok {
			t.Errorf("[%d] set ok = false, want true", i)
		}
	}

	// and delete itmes with odd keys
	for j := 0; j < n/2; j++ {
		i = 2*j + 1
		assertLenEqual("TestOrderedMapBasicFeatures", t, om, n-j)
		ov, ok := om.Delete(i)
		assertLenEqual("TestOrderedMapBasicFeatures", t, om, n-j-1)

		if 2*i != ov {
			t.Errorf("[%d] del val = %v, want %v", i, ov, 2*i)
		}
		if !ok {
			t.Errorf("[%d] del ok = %v, want %v", i, ok, true)
		}

		// deleting again shouldn't change anything
		ov, ok = om.Delete(i)
		assertLenEqual("TestOrderedMapBasicFeatures", t, om, n-j-1)
		if nil != ov {
			t.Errorf("[%d] del val = %v, want %v", i, ov, nil)
		}
		if ok {
			t.Errorf("[%d] del ok = %v, want %v", i, ok, false)
		}
	}

	// get the whole range
	for j := 0; j < n/2; j++ {
		i = 2 * j
		ov, ok := om.Get(i)
		if 4*i != ov {
			t.Errorf("[%d] get val = %v, want %v", i, ov, 4*i)
		}
		if !ok {
			t.Errorf("[%d] gel ok = %v, want %v", i, true, false)
		}

		i = 2*j + 1
		ov, ok = om.Get(i)
		if nil != ov {
			t.Errorf("[%d] gel val = %v, want %v", i, ov, nil)
		}
		if ok {
			t.Errorf("[%d] gel ok = %v, want %v", i, ok, false)
		}
	}

	// check iterations again
	i = 0
	for mi := om.Front(); mi != nil; mi = mi.Next() {
		if i != mi.key {
			t.Errorf("[%d] mi.Key = %v, want %v", i, mi.key, i)
		}
		if i*4 != mi.Value {
			t.Errorf("[%d] mi.Value = %v, want %v", i, mi.Value, i*4)
		}
		i += 2
	}
	i = 2 * ((n - 1) / 2)
	for mi := om.Back(); mi != nil; mi = mi.Prev() {
		if i != mi.key {
			t.Errorf("[%d] mi.Key = %v, want %v", i, mi.key, i)
		}
		if i*4 != mi.Value {
			t.Errorf("[%d] mi.Value = %v, want %v", i, mi.Value, i*4)
		}
		i -= 2
	}
}

func TestOrderedMapUpdatingDoesntChangePairsOrder(t *testing.T) {
	om := NewOrderedMap("foo", "bar", 12, 28, 78, 100, "bar", "baz")

	ov, ok := om.Set(78, 102)
	if ov != 100 {
		t.Errorf("om.Set(78, 102) = %v, want %v", ov, 100)
	}
	if !ok {
		t.Errorf("om.Set(78, 102) = %v, want %v", ok, true)
	}

	assertOrderedPairsEqual(t, om,
		[]interface{}{"foo", 12, 78, "bar"},
		[]interface{}{"bar", 28, 102, "baz"})
}

func TestOrderedMapDeletingAndReinsertingChangesPairsOrder(t *testing.T) {
	om := NewOrderedMap()
	om.Set("foo", "bar")
	om.Set(12, 28)
	om.Set(78, 100)
	om.Set("bar", "baz")

	// delete a item
	ov, ok := om.Delete(78)
	if ov != 100 {
		t.Errorf("om.Delete(78) = %v, want %v", ov, 100)
	}
	if !ok {
		t.Errorf("om.Delete(78) = %v, want %v", ok, true)
	}

	// re-insert the same item
	ov, ok = om.Set(78, 100)
	if ov != nil {
		t.Errorf("om.Delete(78) = %v, want %v", ov, nil)
	}
	if ok {
		t.Errorf("om.Delete(78) = %v, want %v", ok, false)
	}

	assertOrderedPairsEqual(t, om,
		[]interface{}{"foo", 12, "bar", 78},
		[]interface{}{"bar", 28, "baz", 100})
}

func TestOrderedMapEmptyMapOperations(t *testing.T) {
	om := NewOrderedMap()

	var ov interface{}
	var ok bool

	ov, ok = om.Get("foo")
	if ov != nil {
		t.Errorf("om.Get(foo) = %v, want %v", ov, nil)
	}
	if ok {
		t.Errorf("om.Get(foo) = %v, want %v", ok, false)
	}

	ov, ok = om.Delete("bar")
	if ov != nil {
		t.Errorf("om.Delete(bar) = %v, want %v", ov, nil)
	}
	if ok {
		t.Errorf("om.Delete(bar) = %v, want %v", ok, false)
	}

	assertLenEqual("TestOrderedMapEmptyMapOperations", t, om, 0)

	oi := om.Front()
	if oi != nil {
		t.Errorf("om.Front() = %v, want %v", oi, nil)
	}
	oi = om.Back()
	if oi != nil {
		t.Errorf("om.Back() = %v, want %v", oi, nil)
	}
}

type dummyTestStruct struct {
	value string
}

func TestOrderedMapPackUnpackStructs(t *testing.T) {
	om := NewOrderedMap()
	om.Set("foo", dummyTestStruct{"foo!"})
	om.Set("bar", dummyTestStruct{"bar!"})

	ov, ok := om.Get("foo")
	if !ok {
		t.Fatalf(`om.Get("foo") = %v`, ok)
	}
	if "foo!" != ov.(dummyTestStruct).value {
		t.Fatalf(`om.Get("foo") = %v, want %v`, ov, "foo!")
	}

	ov, ok = om.Set("bar", dummyTestStruct{"baz!"})
	if !ok {
		t.Fatalf(`om.Set("bar") = %v`, ok)
	}
	if "bar!" != ov.(dummyTestStruct).value {
		t.Fatalf(`om.Set("bar") = %v, want %v`, ov, "bar!")
	}

	ov, ok = om.Get("bar")
	if !ok {
		t.Fatalf(`om.Get("bar") = %v`, ok)
	}
	if "baz!" != ov.(dummyTestStruct).value {
		t.Fatalf(`om.Get("bar") = %v, want %v`, ov, "baz!")
	}
}

func TestOrderedMapShuffle(t *testing.T) {
	ranLen := 100

	for _, n := range []int{0, 10, 20, 100, 1000, 10000} {
		t.Run(fmt.Sprintf("shuffle test with %d items", n), func(t *testing.T) {
			om := NewOrderedMap()

			keys := make([]interface{}, n)
			values := make([]interface{}, n)

			for i := 0; i < n; i++ {
				// we prefix with the number to ensure that we don't get any duplicates
				keys[i] = fmt.Sprintf("%d_%s", i, randomHexString(t, ranLen))
				values[i] = randomHexString(t, ranLen)

				ov, ok := om.Set(keys[i], values[i])
				if ok {
					t.Fatalf(`[%d] om.Set(%v) = %v`, i, keys[i], ok)
				}
				if ov != nil {
					t.Fatalf(`[%d] om.Set(%v) = %v`, i, keys[i], ov)
				}
			}

			assertOrderedPairsEqual(t, om, keys, values)
		})
	}
}

func TestOrderedMapTemplateRange(t *testing.T) {
	om := NewOrderedMap("z", "Z", "a", "A")
	tmpl, err := template.New("test").Parse("{{range $e := .om.Items}}[ {{$e.Key}} = {{$e.Value}} ]{{end}}")
	if err != nil {
		t.Fatal(err.Error())
	}

	cm := map[string]interface{}{
		"om": om,
	}
	sb := &strings.Builder{}
	err = tmpl.Execute(sb, cm)
	if err != nil {
		t.Fatal(err.Error())
	}

	a := sb.String()
	w := "[ z = Z ][ a = A ]"
	if w != a {
		t.Errorf("tmpl.Execute() = %q, want %q", a, w)
	}
}

/* Test helpers */
func assertOrderedPairsEqual(t *testing.T, om *OrderedMap, eks, evs []interface{}) {
	assertOrderedPairsEqualFromNewest(t, om, eks, evs)
	assertOrderedPairsEqualFromOldest(t, om, eks, evs)
}

func assertOrderedPairsEqualFromNewest(t *testing.T, om *OrderedMap, eks, evs []interface{}) {
	if len(eks) != len(evs) {
		t.Errorf("len(keys) %v != len(vals) %v", len(eks), len(evs))
		return
	}

	if len(eks) != om.Len() {
		t.Errorf("len(keys) %v != om.Len %v", len(eks), om.Len())
		return
	}

	i := om.Len() - 1
	for item := om.Back(); item != nil; item = item.Prev() {
		if eks[i] != item.key {
			t.Errorf("[%d] key = %v, want %v", i, item.key, eks[i])
		}

		if evs[i] != item.Value {
			t.Errorf("[%d] val = %v, want %v", i, item.Value, evs[i])
		}
		i--
	}
}

func assertOrderedPairsEqualFromOldest(t *testing.T, om *OrderedMap, eks, evs []interface{}) {
	if len(eks) != len(evs) {
		t.Errorf("len(keys) %v != len(vals) %v", len(eks), len(evs))
		return
	}

	if len(eks) != om.Len() {
		t.Errorf("len(keys) %v != om.Len %v", len(eks), om.Len())
		return
	}

	i := 0
	for item := om.Front(); item != nil; item = item.Next() {
		if eks[i] != item.key {
			t.Errorf("[%d] key = %v, want %v", i, item.key, eks[i])
		}

		if evs[i] != item.Value {
			t.Errorf("[%d] val = %v, want %v", i, item.Value, evs[i])
		}
		i++
	}
}

func assertLenEqual(n string, t *testing.T, om *OrderedMap, w int) {
	if om.Len() != w {
		t.Errorf("%s: om.Len() != %v", n, w)
	}
	if om.list.Len() != w {
		t.Errorf("%s: om.list.Len() != %v", n, w)
	}
}

func randomHexString(t *testing.T, length int) string {
	b := length / 2
	randBytes := make([]byte, b)

	if n, err := rand.Read(randBytes); err != nil || n != b {
		if err == nil {
			err = fmt.Errorf("only got %v random bytes, expected %v", n, b)
		}
		t.Fatal(err)
	}

	return hex.EncodeToString(randBytes)
}

func TestOrderedMapString(t *testing.T) {
	w := `{"1":1,"3":3,"2":2}`
	a := fmt.Sprintf("%s", NewOrderedMap("1", 1, "3", 3, "2", 2))
	if w != a {
		t.Errorf("TestOrderedMapString = %v, want %v", a, w)
	}
}

/*----------- JOSN Test -----------------*/
func TestOrderedMapMarshal(t *testing.T) {
	om := NewOrderedMap()
	om.Set("a", 34)
	om.Set("b", []int{3, 4, 5})
	b, err := json.Marshal(om)
	if err != nil {
		t.Fatalf("Marshal OrderedMap: %v", err)
	}
	// fmt.Printf("%q\n", b)
	const expected = "{\"a\":34,\"b\":[3,4,5]}"
	if !bytes.Equal(b, []byte(expected)) {
		t.Errorf("Marshal OrderedMap: %q not equal to expected %q", b, expected)
	}
}

func ExampleOrderedMap_UnmarshalJSON() {
	const jsonStream = `{
  "country"     : "United States",
  "countryCode" : "US",
  "region"      : "CA",
  "regionName"  : "California",
  "city"        : "Mountain View",
  "zip"         : "94043",
  "lat"         : 37.4192,
  "lon"         : -122.0574,
  "timezone"    : "America/Los_Angeles",
  "isp"         : "Google Cloud",
  "org"         : "Google Cloud",
  "as"          : "AS15169 Google Inc.",
  "mobile"      : true,
  "proxy"       : false,
  "query"       : "35.192.xx.xxx"
}`

	// compare with if using a regular generic map, the unmarshalled result
	//  is a map with unpredictable order of keys
	var m map[string]interface{}
	err := json.Unmarshal([]byte(jsonStream), &m)
	if err != nil {
		fmt.Println("error:", err)
	}
	for key := range m {
		// fmt.Printf("%-12s: %v\n", key, m[key])
		_ = key
	}

	// use the OrderedMap to Unmarshal from JSON object
	om := NewOrderedMap()
	err = json.Unmarshal([]byte(jsonStream), om)
	if err != nil {
		fmt.Println("error:", err)
	}

	// loop over all key-value pairs,
	// it is ok to call Set append-modify new key-value pairs,
	// but not safe to call Delete during iteration.
	for me := om.Front(); me != nil; me = me.Next() {
		fmt.Printf("%-12s: %v\n", me.Key(), me.Value)
		if me.Key() == "city" {
			om.Set("mobile", false)
			om.Set("extra", 42)
		}
	}

	// Output:
	// country     : United States
	// countryCode : US
	// region      : CA
	// regionName  : California
	// city        : Mountain View
	// zip         : 94043
	// lat         : 37.4192
	// lon         : -122.0574
	// timezone    : America/Los_Angeles
	// isp         : Google Cloud
	// org         : Google Cloud
	// as          : AS15169 Google Inc.
	// mobile      : false
	// proxy       : false
	// query       : 35.192.xx.xxx
	// extra       : 42
}

func TestOrderedMapUnmarshalFromInvalid(t *testing.T) {
	om := NewOrderedMap()

	om.Set("m", math.NaN())
	b, err := json.Marshal(om)
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", b, err)
	}
	// fmt.Println(om, b, err)
	om.Delete("m")

	err = json.Unmarshal([]byte("[]"), om)
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error")
	}

	err = json.Unmarshal([]byte("["), om)
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", om)
	}

	err = om.UnmarshalJSON([]byte(nil))
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", om)
	}

	err = om.UnmarshalJSON([]byte("{}3"))
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", om)
	}

	err = om.UnmarshalJSON([]byte("{"))
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", om)
	}

	err = om.UnmarshalJSON([]byte("{]"))
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", om)
	}

	err = om.UnmarshalJSON([]byte(`{"a": 3, "b": [{`))
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", om)
	}

	err = om.UnmarshalJSON([]byte(`{"a": 3, "b": [}`))
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error:", om)
	}
	// fmt.Println("error:", om, err)
}

func TestOrderedMapUnmarshal(t *testing.T) {
	var (
		data  = []byte(`{"as":"AS15169 Google Inc.","city":"Mountain View","country":"United States","countryCode":"US","isp":"Google Cloud","lat":37.4192,"lon":-122.0574,"org":"Google Cloud","query":"35.192.25.53","region":"CA","regionName":"California","status":"success","timezone":"America/Los_Angeles","zip":"94043"}`)
		pairs = []interface{}{
			"as", "AS15169 Google Inc.",
			"city", "Mountain View",
			"country", "United States",
			"countryCode", "US",
			"isp", "Google Cloud",
			"lat", 37.4192,
			"lon", -122.0574,
			"org", "Google Cloud",
			"query", "35.192.25.53",
			"region", "CA",
			"regionName", "California",
			"status", "success",
			"timezone", "America/Los_Angeles",
			"zip", "94043",
		}
		obj = NewOrderedMap(pairs...)
	)

	om := NewOrderedMap()
	err := json.Unmarshal(data, om)
	if err != nil {
		t.Fatalf("Unmarshal OrderedMap: %v", err)
	}

	// check by Has and GetValue
	for i := 0; i+1 < len(pairs); i += 2 {
		k := pairs[i]
		v := pairs[i+1]

		if !om.Has(k) {
			t.Fatalf("expect key %q exists in Unmarshaled OrderedMap", k)
		}
		value, ok := om.Get(k)
		if !ok || value != v {
			t.Fatalf("expect for key %q: the value %v should equal to %v, in Unmarshaled OrderedMap", k, value, v)
		}
	}

	b, err := json.MarshalIndent(om, "", "  ")
	if err != nil {
		t.Fatalf("Unmarshal OrderedMap: %v", err)
	}
	const expected = `{
  "as": "AS15169 Google Inc.",
  "city": "Mountain View",
  "country": "United States",
  "countryCode": "US",
  "isp": "Google Cloud",
  "lat": 37.4192,
  "lon": -122.0574,
  "org": "Google Cloud",
  "query": "35.192.25.53",
  "region": "CA",
  "regionName": "California",
  "status": "success",
  "timezone": "America/Los_Angeles",
  "zip": "94043"
}`
	if !bytes.Equal(b, []byte(expected)) {
		t.Fatalf("Unmarshal OrderedMap marshal indent from %#v not equal to expected: %q\n", om, expected)
	}

	if !reflect.DeepEqual(om, obj) {
		t.Fatalf("Unmarshal OrderedMap not deeply equal: %#v %#v", om, obj)
	}

	val, ok := om.Delete("org")
	if !ok {
		t.Fatalf("org should exist")
	}
	om.Set("org", val)
	b, err = json.MarshalIndent(om, "", "  ")
	// fmt.Println("after delete", om, string(b), err)
	if err != nil {
		t.Fatalf("Unmarshal OrderedMap: %v", err)
	}
	const expected2 = `{
  "as": "AS15169 Google Inc.",
  "city": "Mountain View",
  "country": "United States",
  "countryCode": "US",
  "isp": "Google Cloud",
  "lat": 37.4192,
  "lon": -122.0574,
  "query": "35.192.25.53",
  "region": "CA",
  "regionName": "California",
  "status": "success",
  "timezone": "America/Los_Angeles",
  "zip": "94043",
  "org": "Google Cloud"
}`
	if !bytes.Equal(b, []byte(expected2)) {
		t.Fatalf("Unmarshal OrderedMap marshal indent from %#v not equal to expected: %s\n", om, expected2)
	}
}

func TestOrderedMapUnmarshalNested(t *testing.T) {
	var (
		data = []byte(`{"a": true, "b": [3, 4, { "b": "3", "d": [] }]}`)
		obj  = NewOrderedMap(
			"a", true,
			"b", JSONArray{float64(3), float64(4), NewOrderedMap("b", "3", "d", JSONArray{})},
		)
	)

	om := NewOrderedMap()
	err := json.Unmarshal(data, om)
	if err != nil {
		t.Fatalf("Unmarshal OrderedMap: %v", err)
	}

	if !reflect.DeepEqual(om, obj) {
		t.Fatalf("Unmarshal OrderedMap not deeply equal: %#v expected %#v", om, obj)
	}
}

func ExampleNewOrderedMap() {
	// initialize from a list of key-value pairs
	om := NewOrderedMap(
		"country", "United States",
		"countryCode", "US",
		"region", "CA",
		"regionName", "California",
		"city", "Mountain View",
		"zip", "94043",
		"lat", 37.4192,
		"lon", -122.0574,
		"timezone", "America/Los_Angeles",
		"isp", "Google Cloud",
		"org", "Google Cloud",
		"as", "AS15169 Google Inc.",
		"mobile", true,
		"proxy", false,
		"query", "35.192.xx.xxx",
	)

	for me := om.Back(); me != nil; me = me.Prev() {
		fmt.Printf("%-12s: %v\n", me.Key(), me.Value)
	}

	// Output:
	// query       : 35.192.xx.xxx
	// proxy       : false
	// mobile      : true
	// as          : AS15169 Google Inc.
	// org         : Google Cloud
	// isp         : Google Cloud
	// timezone    : America/Los_Angeles
	// lon         : -122.0574
	// lat         : 37.4192
	// zip         : 94043
	// city        : Mountain View
	// regionName  : California
	// region      : CA
	// countryCode : US
	// country     : United States
}

var unmarshalTests = []struct {
	in  string
	out interface{}
	err interface{}
}{
	{in: "{}", out: NewOrderedMap()},
	{in: `{"a": 3}`, out: NewOrderedMap("a", float64(3))},
	{in: `{"a": 3, "b": true}`, out: NewOrderedMap("a", float64(3), "b", true)},
	{in: `{"a": 3, "b": true, "c": null}`, out: NewOrderedMap("a", float64(3), "b", true, "c", nil)},
	{in: `{"a": 3, "c": null, "d": []}`, out: NewOrderedMap("a", float64(3), "c", nil, "d", JSONArray{})},
	{in: `{"a": 3, "c": null, "d": [3,4,true]}`, out: NewOrderedMap(
		"a", float64(3), "c", nil, "d", JSONArray{
			float64(3), float64(4), true,
		})},
	{in: `{"a": 3, "c": null, "d": [3,4,true, { "inner": "abc" }]}`, out: NewOrderedMap(
		"a", float64(3), "c", nil, "d", JSONArray([]interface{}{
			float64(3), float64(4), true, NewOrderedMap("inner", "abc"),
		}))},
}

func TestOrderedMapUnmarshals(t *testing.T) {
	for i, tt := range unmarshalTests {
		in := []byte(tt.in)

		v := NewOrderedMap()
		dec := json.NewDecoder(bytes.NewReader(in))
		if err := dec.Decode(v); !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%d: %v, want %v", i, err, tt.err)
			continue
		} else if err != nil {
			continue
		}

		if !reflect.DeepEqual(v, tt.out) {
			act, _ := json.Marshal(v)
			exp, _ := json.Marshal(tt.out)
			t.Errorf("#%d: mismatch\nhave: %#+v\nwant: %#+v\nact: %s\nexp:%s", i, v, tt.out, string(act), string(exp))
			continue
		}

		// Check round trip also decodes correctly.
		if tt.err == nil {
			enc, err := json.Marshal(v)
			if err != nil {
				t.Errorf("#%d: error re-marshaling: %v", i, err)
				continue
			}

			vv := NewOrderedMap()
			dec = json.NewDecoder(bytes.NewReader(enc))
			if err := dec.Decode(vv); err != nil {
				t.Errorf("#%d: error re-unmarshaling %#q: %v", i, enc, err)
				continue
			}
			if !reflect.DeepEqual(v, vv) {
				t.Errorf("#%d: mismatch\nhave: %#+v\nwant: %#+v", i, v, vv)
				t.Errorf("     In: %q", strings.Map(noSpace, string(in)))
				t.Errorf("Marshal: %q", strings.Map(noSpace, string(enc)))
				continue
			}
		}
	}
}

func noSpace(c rune) rune {
	if isSpace(byte(c)) { //only used for ascii
		return -1
	}
	return c
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
