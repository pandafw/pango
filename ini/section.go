package ini

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/pandafw/pango/col"
	"github.com/pandafw/pango/iox"
)

// Entry ini entry
type Entry struct {
	Value    string
	Comments []string
}

// Section ini section
type Section struct {
	name     string          // Name for tihs section.
	comments []string        // Comment for this section.
	entries  *col.OrderedMap // Entries for this section.
}

// NewSection create a INI section
func NewSection(name string, comments ...string) *Section {
	return &Section{
		name:     name,
		comments: comments,
		entries:  col.NewOrderedMap(),
	}
}

// Name return the section's name
func (sec *Section) Name() string {
	return sec.name
}

// Comments return the section's comment string array
func (sec *Section) Comments() []string {
	return sec.comments
}

// Keys return the section's key string array
func (sec *Section) Keys() []string {
	ks := make([]string, 0, sec.entries.Len())
	for e := sec.entries.Front(); e != nil; e = e.Next() {
		ks = append(ks, e.Key().(string))
	}
	return ks
}

// StringMap return the section's entries key.(string)/value.(string) map
func (sec *Section) StringMap() map[string]string {
	m := make(map[string]string, sec.entries.Len())
	for e := sec.entries.Front(); e != nil; e = e.Next() {
		var v string
		switch se := e.Value.(type) {
		case *col.List:
			v = se.Front().Value.(string)
		case *Entry:
			v = se.Value
		}
		m[e.Key().(string)] = v
	}
	return m
}

// StringsMap return the section's entries key.(string)/value.([]string) map
func (sec *Section) StringsMap() map[string][]string {
	m := make(map[string][]string, sec.entries.Len())
	for e := sec.entries.Front(); e != nil; e = e.Next() {
		var v []string
		switch se := e.Value.(type) {
		case *col.List:
			v = sec.toStrings(se)
		case *Entry:
			v = []string{se.Value}
		}
		m[e.Key().(string)] = v
	}
	return m
}

// Map return the section's entries key.(string)/value.(interface{}) map
func (sec *Section) Map() map[string]interface{} {
	m := make(map[string]interface{}, sec.entries.Len())
	for e := sec.entries.Front(); e != nil; e = e.Next() {
		var v interface{}
		switch se := e.Value.(type) {
		case *col.List:
			v = sec.toStrings(se)
		case *Entry:
			v = se.Value
		}
		m[e.Key().(string)] = v
	}
	return m
}

// Add add a key/value entry to the section
func (sec *Section) Add(key string, value string, comments ...string) *Entry {
	e := &Entry{Value: value, Comments: comments}
	sec.add(key, e)
	return e
}

// add add a key/value entry to the section
func (sec *Section) add(key string, e *Entry) {
	if v, ok := sec.entries.Get(key); ok {
		if l, ok := v.(*col.List); ok {
			l.PushBack(e)
			return
		}
		l := col.NewList()
		l.PushBack(v)
		l.PushBack(e)
		sec.entries.Set(key, l)
		return
	}

	sec.entries.Set(key, e)
}

// Set set a key/value entry to the section
func (sec *Section) Set(key string, value string, comments ...string) *Entry {
	e := &Entry{Value: value, Comments: comments}
	sec.entries.Set(key, e)
	return e
}

// Get get a value of the key from the section
func (sec *Section) Get(key string) string {
	e := sec.GetEntry(key)
	if e != nil {
		return e.Value
	}
	return ""
}

// GetString get a string value of the key from the section
// if not found, returns the default defs[0] string value
func (sec *Section) GetString(key string, defs ...string) string {
	e := sec.GetEntry(key)
	if e != nil {
		return e.Value
	}
	if len(defs) > 0 {
		return defs[0]
	}
	return ""
}

// GetInt get a int value of the key from the section
// if not found or convert error, returns the default defs[0] int value
func (sec *Section) GetInt(key string, defs ...int) int {
	e := sec.GetEntry(key)
	if e != nil {
		if i, err := strconv.ParseInt(e.Value, 0, strconv.IntSize); err == nil {
			return int(i)
		}
	}
	if len(defs) > 0 {
		return defs[0]
	}
	return 0
}

// GetFloat get a float value of the key from the section
// if not found or convert error, returns the default defs[0] float value
func (sec *Section) GetFloat(key string, defs ...float64) float64 {
	e := sec.GetEntry(key)
	if e != nil {
		if f, err := strconv.ParseFloat(e.Value, 0); err == nil {
			return f
		}
	}
	if len(defs) > 0 {
		return defs[0]
	}
	return 0
}

// GetBool get a bool value of the key from the section
// if not found, returns the default defs[0] int value
func (sec *Section) GetBool(key string, defs ...bool) bool {
	e := sec.GetEntry(key)
	if e != nil {
		if b, err := strconv.ParseBool(e.Value); err == nil {
			return b
		}
	}
	if len(defs) > 0 {
		return defs[0]
	}
	return false
}

func (sec *Section) toStrings(l *col.List) []string {
	ss := make([]string, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		ss = append(ss, e.Value.(*Entry).Value)
	}
	return ss
}

// GetValues get the key's values from the section
func (sec *Section) GetValues(key string) []string {
	if v, ok := sec.entries.Get(key); ok {
		switch se := v.(type) {
		case *col.List:
			return sec.toStrings(se)
		case *Entry:
			return []string{se.Value}
		}
	}
	return nil
}

// GetEntry get the key's entry from the section
func (sec *Section) GetEntry(key string) *Entry {
	if v, ok := sec.entries.Get(key); ok {
		switch se := v.(type) {
		case *col.List:
			return se.Front().Value.(*Entry)
		case *Entry:
			return se
		}
	}
	return nil
}

// Clear clear the entries and comments
func (sec *Section) Clear() {
	sec.comments = nil
	sec.entries.Clear()
}

// Copy copy entries from src section, overrite existing entries
func (sec *Section) Copy(src *Section) {
	if len(src.comments) > 0 {
		sec.comments = src.comments
	}
	sec.entries.Copy(src.entries)
}

// Merge merge entries from src section
func (sec *Section) Merge(src *Section) {
	sec.comments = append(sec.comments, src.comments...)
	for e := src.entries.Front(); e != nil; e = e.Next() {
		sec.add(e.Key().(string), e.Value.(*Entry))
	}
}

// String write section to string
func (sec *Section) String() string {
	sb := &strings.Builder{}
	bw := bufio.NewWriter(sb)
	sec.Write(bw, iox.EOL)
	bw.Flush()
	return sb.String()
}

// Write output section to the writer
func (sec *Section) Write(bw *bufio.Writer, eol string) error {
	// comments
	if err := sec.writeComments(bw, sec.comments, eol); err != nil {
		return err
	}

	// section name
	if err := sec.writeSectionName(bw, eol); err != nil {
		return err
	}

	// section entries
	if err := sec.writeSectionEntries(bw, eol); err != nil {
		return err
	}

	// blank line
	if _, err := bw.WriteString(eol); err != nil {
		return err
	}

	return nil
}

func (sec *Section) writeComments(bw *bufio.Writer, comments []string, eol string) (err error) {
	for _, s := range comments {
		_, err = bw.WriteString(s)
		_, err = bw.WriteString(eol)
	}
	return err
}

func (sec *Section) writeSectionName(bw *bufio.Writer, eol string) (err error) {
	if sec.name != "" {
		err = bw.WriteByte('[')
		_, err = bw.WriteString(sec.name)
		err = bw.WriteByte(']')
	}
	_, err = bw.WriteString(eol)
	return err
}

func (sec *Section) writeSectionEntries(bw *bufio.Writer, eol string) (err error) {
	for me := sec.entries.Front(); me != nil; me = me.Next() {
		switch se := me.Value.(type) {
		case *col.List:
			for le := se.Front(); le != nil; le = le.Next() {
				if err = sec.writeSectionEntry(bw, me.Key().(string), le.Value.(*Entry), eol); err != nil {
					return err
				}
			}
		case *Entry:
			if err = sec.writeSectionEntry(bw, me.Key().(string), se, eol); err != nil {
				return err
			}
		}
	}
	return err
}

func (sec *Section) writeSectionEntry(bw *bufio.Writer, key string, ve *Entry, eol string) (err error) {
	if len(ve.Comments) > 0 {
		_, err = bw.WriteString(eol)
		err = sec.writeComments(bw, ve.Comments, eol)
	}

	_, err = bw.WriteString(key)
	err = bw.WriteByte(' ')
	err = bw.WriteByte('=')
	err = bw.WriteByte(' ')
	_, err = bw.WriteString(quote(ve.Value))
	_, err = bw.WriteString(eol)
	return err
}

func (sec *Section) writeKeyValue(bw *bufio.Writer, key string, val string, eol string) (err error) {
	_, err = bw.WriteString(key)
	err = bw.WriteByte(' ')
	err = bw.WriteByte('=')
	err = bw.WriteByte(' ')
	_, err = bw.WriteString(quote(val))
	_, err = bw.WriteString(eol)
	return err
}
