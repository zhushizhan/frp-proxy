//go:build !noweb && !nowebui

package frpc

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var embedFS embed.FS

func HTTPFileSystem() (http.FileSystem, bool) {
	subFS, err := fs.Sub(embedFS, "dist")
	if err != nil {
		return nil, false
	}
	return http.FS(subFS), true
}
