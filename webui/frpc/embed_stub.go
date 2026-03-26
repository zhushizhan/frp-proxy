//go:build noweb || nowebui

package frpc

import "net/http"

func HTTPFileSystem() (http.FileSystem, bool) {
	return nil, false
}
