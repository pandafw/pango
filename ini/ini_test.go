package ini

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/pandafw/pango/iox"
)

func TestLoadFile(t *testing.T) {
	fin := "testdata/input.ini"
	fexp := "testdata/except.ini"
	fout := "testdata/output.ini"

	defer os.Remove(fout)

	ini := NewIni()
	ini.EOL = iox.CRLF
	ini.Multiple = true

	// load
	if ini.LoadFile(fin) != nil {
		t.Errorf("ini.LoadFile(%q) != nil", fin)
	}

	// value
	other := ini.Section("other")
	if other == nil {
		t.Error(`ini.Section("other") == nil`)
		return
	}

	dec := other.GetInt("dec")
	if 42 != dec {
		t.Errorf(`other.GetInt("dec") = %v, want %v`, dec, 42)
	}
	hex := other.GetInt("hex")
	if 42 != hex {
		t.Errorf(`other.GetInt("hex") = %v, want %v`, dec, 42)
	}
	vtrue := other.GetBool("true")
	if !vtrue {
		t.Error(`other.GetBool("true") != true`)
	}
	vfalse := other.GetBool("false")
	if vfalse {
		t.Error(`other.GetBool("false") != false`)
	}

	// expected file
	bexp, _ := ioutil.ReadFile(fexp)
	sexp := string(bexp)

	// write data
	{
		sout := &strings.Builder{}
		werr := ini.WriteData(sout)
		if werr != nil {
			t.Errorf(`ini.WriteData(sout) = %v`, werr)
		}
		if sexp != sout.String() {
			t.Errorf(`ini.WriteData(sout)\n actual: %v\n   want: %v`, sout.String(), sexp)
		}
	}

	// write file
	{
		werr := ini.WriteFile(fout)
		if werr != nil {
			t.Errorf(`ini.WriteFile(fout) = %v`, werr)
		}

		bout, _ := ioutil.ReadFile(fout)
		sout := string(bout)
		if sexp != sout {
			t.Errorf(`ini.WriteFile(fout)\n actual: %v\n   want: %v`, fout, sexp)
		}
	}

	// remove section
	{
		sec := ini.RemoveSection("")
		if sec == nil {
			t.Errorf(`ini.RemoveSection("") = %v`, sec)
		}
		sec = ini.RemoveSection("not exist")
		if sec != nil {
			t.Errorf(`ini.RemoveSection("not exist") = %v`, sec)
		}
	}
}
