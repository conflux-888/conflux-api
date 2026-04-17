package main

import (
	"embed"
	"io/fs"
)

//go:embed all:admin/dist
var adminDistFS embed.FS

// adminFS returns the embedded admin UI files with the admin/dist prefix stripped.
func adminFS() fs.FS {
	sub, _ := fs.Sub(adminDistFS, "admin/dist")
	return sub
}
