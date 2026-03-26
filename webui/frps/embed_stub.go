//go:build noweb || nowebui

package frps

import "net/http"

func HTTPFileSystem() (http.FileSystem, bool) {
	return nil, false
}
