package tpl

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// HTMLTemplate html template engine
type HTMLTemplate struct {
	Extension string  // template extension
	Funcs     FuncMap // template functions
	Delims    Delims  // delimeters

	template *template.Template
}

// NewHTMLTemplate new template engine
func NewHTMLTemplate() *HTMLTemplate {
	return &HTMLTemplate{
		Extension: ".html",
		Delims:    Delims{Left: "{{", Right: "}}"},
		template:  template.New(""),
	}
}

// Load glob and parse template files under root path
func (ht *HTMLTemplate) Load(root string) error {
	tpl := template.New("")
	tpl.Funcs(template.FuncMap(ht.Funcs))

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		return ht.loadFile(tpl, nil, root, path)
	})

	if err != nil {
		return err
	}

	ht.template = tpl
	return nil
}

// LoadFS glob and parse template files from FS
func (ht *HTMLTemplate) LoadFS(fsys fs.FS, root string) error {
	tpl := template.New("")
	tpl.Funcs(template.FuncMap(ht.Funcs))

	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		return ht.loadFile(tpl, fsys, root, path)
	})

	if err != nil {
		return err
	}

	ht.template = tpl
	return nil
}

// loadFile load template file
func (ht *HTMLTemplate) loadFile(tpl *template.Template, fsys fs.FS, root, path string) error {
	if filepath.Ext(path) != ht.Extension {
		return nil
	}

	text, err := readFile(fsys, path)
	if err != nil {
		return fmt.Errorf("HTMLTemplate load template %q error: %v", path, err)
	}

	path = toTemplateName(root, path, ht.Extension)

	tpl = tpl.New(path)
	_, err = tpl.Parse(text)
	if err != nil {
		return fmt.Errorf("HTMLTemplate parse template %q error: %v", path, err)
	}
	return nil
}

// Render render template with io.Writer
func (ht *HTMLTemplate) Render(w io.Writer, name string, data interface{}) error {
	err := ht.template.ExecuteTemplate(w, name, data)
	if err != nil {
		return fmt.Errorf("HTMLTemplate execute template %q error: %v", name, err)
	}

	return nil
}
