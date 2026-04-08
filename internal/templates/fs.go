package templates

import "embed"

// FS is an embedded filesystem containing all email templates.
// It includes HTML templates for the various AKS event types.
//
//go:embed *.gohtml
var FS embed.FS
