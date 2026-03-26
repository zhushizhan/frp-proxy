//go:build !noweb && !nowebui

package frpc

import (
	"embed"
	"io/fs"
	"net/http"
)

// Embed the built webui/frpc distribution.
// Keep this file touched when we need to force-refresh embedded client webui assets on Windows builds.
// Client settings UI marker.
//go:embed dist
var embedFS embed.FS

func HTTPFileSystem() (http.FileSystem, bool) {
	subFS, err := fs.Sub(embedFS, "dist")
	if err != nil {
		return nil, false
	}
	return http.FS(subFS), true
}
