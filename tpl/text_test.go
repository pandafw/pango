package tpl

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pandafw/pango/str"
)

func textTestLoad(t *testing.T, tt *TextTemplate) {
	sb := &strings.Builder{}

	ctx := map[string]interface{}{
		"Title":   "Front Page",
		"Message": "Hello world!",
		"Time":    time.Now(),
	}
	err := tt.Render(sb, "index", ctx)
	if err != nil {
		t.Errorf(`tt.Render(sb, "index", ctx) = %v`, err)
		return
	}

	txt := sb.String()
	if !str.Contains(txt, fmt.Sprintf("<title>%s</title>", ctx["Title"])) {
		t.Errorf("Incorrect Title\n%s", txt)
	}
	if !str.Contains(txt, fmt.Sprintf("<p>%s</p>", ctx["Message"])) {
		t.Errorf("Incorrect Message\n%s", txt)
	}
	if !str.Contains(txt, fmt.Sprintf("<p>Time: %s</p>", ctx["Time"].(time.Time).Format("2006/1/2 15:04:05"))) {
		t.Errorf("Incorrect Message\n%s", txt)
	}

	sb.Reset()
	ctx = map[string]interface{}{
		"Title":   "Admin Page",
		"Message": "Hello world!",
		"Time":    time.Now(),
	}
	err = tt.Render(sb, "admin/admin", ctx)
	if err != nil {
		t.Errorf(`tt.Render(sb, "admin/admin", ctx) = %v`, err)
		return
	}
	txt = sb.String()
	if !str.Contains(txt, fmt.Sprintf("<title>%s</title>", ctx["Title"])) {
		t.Errorf("Incorrect Title\n%s", txt)
	}
	if !str.Contains(txt, fmt.Sprintf("<p>Admin: %s</p>", ctx["Message"])) {
		t.Errorf("Incorrect Message\n%s", txt)
	}
	if !str.Contains(txt, fmt.Sprintf("<p>Time: %s</p>", ctx["Time"].(time.Time).Format("2006/1/2 15:04:05"))) {
		t.Errorf("Incorrect Message\n%s", txt)
	}
}

func TestLoadText(t *testing.T) {
	tt := NewTextTemplate()
	root := "testdata"

	err := tt.Load(root)
	if err != nil {
		t.Errorf(`ht.Load(%q) = %v`, root, err)
		return
	}

	textTestLoad(t, tt)
}

func TestFSLoadText(t *testing.T) {
	tt := NewTextTemplate()
	root := "testdata"

	err := tt.LoadFS(testdata, root)
	if err != nil {
		t.Errorf(`ht.LoadFS(%q) = %v`, root, err)
		return
	}

	textTestLoad(t, tt)
}
