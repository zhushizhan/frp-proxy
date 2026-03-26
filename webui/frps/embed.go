//go:build !noweb && !nowebui

package frps

import (
	"embed"
	"io/fs"
	"net/http"
)

// Embed the built webui/frps distribution.
// Keep this file touched when we need to force-refresh embedded webui assets on Windows builds.
// Layout refresh marker: common vs advanced settings separation.
//go:embed dist
var embedFS embed.FS

func HTTPFileSystem() (http.FileSystem, bool) {
	subFS, err := fs.Sub(embedFS, "dist")
	if err != nil {
		return nil, false
	}
	return http.FS(subFS), true
}
