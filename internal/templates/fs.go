package templates

import (
	"embed"
	"errors"
	"html/template"
)

// FS is an embedded filesystem containing all email templates.
// It includes HTML templates for the various AKS event types.
//
//go:embed *.gohtml
var FS embed.FS

// MustBuild builds a map of template names to parsed templates from the embedded filesystem.
// It panics if any template fails to parse.
func MustBuild(tmplFiles ...string) map[string]*template.Template {
	tmpls, err := Build(tmplFiles...)
	if err != nil {
		panic(err)
	}
	return tmpls
}

// Build builds a map of template names to parsed templates from the embedded filesystem.
// It returns an error that wraps all parsing errors.
func Build(tmplFiles ...string) (map[string]*template.Template, error) {
	tmpls := make(map[string]*template.Template, len(tmplFiles))
	var allErrs error
	for _, filename := range tmplFiles {
		var err error
		if tmpls[filename], err = template.ParseFS(FS, filename); err != nil {
			allErrs = errors.Join(err, allErrs)
		}
	}
	return tmpls, allErrs
}
