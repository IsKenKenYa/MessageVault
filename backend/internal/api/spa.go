package api

import (
	"net/http"
	"os"
	urlpath "path"
	"path/filepath"
	"strings"
)

func spaHandler(root string) http.Handler {
	fileServer := http.FileServer(http.Dir(root))
	indexPath := ensureWebRootIndex(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		cleanPath := strings.TrimPrefix(urlpath.Clean("/"+r.URL.Path), "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}
		targetPath := filepath.Join(root, filepath.FromSlash(cleanPath))
		if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}
