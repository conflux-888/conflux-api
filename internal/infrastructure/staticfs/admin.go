package staticfs

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Register mounts a filesystem as a SPA under mountPath (e.g. "/admin").
// Assets are served from their actual paths; all other paths fall back to index.html
// so React Router client-side routes work on direct navigation.
func Register(r *gin.Engine, fsys fs.FS, mountPath string) {
	if fsys == nil {
		return
	}

	fileServer := http.FileServer(http.FS(fsys))

	// Redirect /admin → /admin/
	r.GET(mountPath, func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, mountPath+"/")
	})

	// Catch-all under the mount path
	r.GET(mountPath+"/*filepath", func(c *gin.Context) {
		rel := strings.TrimPrefix(c.Param("filepath"), "/")
		if rel == "" {
			serveIndex(c, fsys)
			return
		}

		// Try to serve the actual file (e.g. assets/*, favicon.ico)
		if f, err := fs.Stat(fsys, rel); err == nil && !f.IsDir() {
			c.Request.URL.Path = "/" + rel
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// SPA fallback
		serveIndex(c, fsys)
	})
}

func serveIndex(c *gin.Context, fsys fs.FS) {
	data, err := fs.ReadFile(fsys, "index.html")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}
