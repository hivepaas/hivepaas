package assets

import (
	"embed"
	"io/fs"
)

//go:embed icons/*.svg
var iconEmbedFS embed.FS

//go:embed templates
var templateEmbedFS embed.FS

// GetIconsFS returns a filesystem scoped to the icons directory
func GetIconsFS() fs.FS {
	subFS, err := fs.Sub(iconEmbedFS, "icons")
	if err != nil {
		panic(err)
	}
	return subFS
}

// GetTemplatesFS returns a filesystem scoped to the templates directory
func GetTemplatesFS() fs.FS {
	subFS, err := fs.Sub(templateEmbedFS, "templates")
	if err != nil {
		panic(err)
	}
	return subFS
}
