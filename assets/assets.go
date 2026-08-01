package assets

import (
	"embed"
	"io/fs"
)

//go:embed icons/*.svg
var iconEmbedFS embed.FS

// GetIconsFS returns a filesystem scoped to the icons directory
func GetIconsFS() fs.FS {
	subFS, err := fs.Sub(iconEmbedFS, "icons")
	if err != nil {
		panic(err)
	}
	return subFS
}
