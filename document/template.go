package document

import (
	"embed"
	"io"
	"io/fs"
	"text/template"
)

//go:embed resources/*
var resources embed.FS

func generateDocument(out io.Writer, s []Section) error {

	err := renderTemplate(resources, "resources/document.md", s, out)
	if err != nil {
		return err
	}

	return nil
}

func renderTemplate(fs fs.FS, tpl string, args interface{}, out io.Writer) error {
	// read the template
	t, err := template.ParseFS(fs, tpl)
	if err != nil {
		return err
	}

	// we render the template
	err = t.Execute(out, args)
	if err != nil {
		return err
	}

	return nil
}
