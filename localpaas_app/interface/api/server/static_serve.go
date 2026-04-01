package server

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	notStaticPrefixes = []string{"/_/"}
)

const INDEX = "index.html"

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
	CacheControlHeader() string
}

// StaticServe returns a middleware handler that serves static files in the given directory.
func StaticServe(urlPrefix string, fs ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if fs.Exists(urlPrefix, c.Request.URL.Path) {
			if cacheHeader := fs.CacheControlHeader(); cacheHeader != "" {
				c.Writer.Header().Set("Cache-Control", cacheHeader)
			}
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

type localFileSystem struct {
	http.FileSystem
	root         string
	indexes      bool
	cacheControl string
}

func localFile(root string, indexes bool, cacheControl string) *localFileSystem {
	return &localFileSystem{
		FileSystem:   gin.Dir(root, indexes),
		root:         root,
		indexes:      indexes,
		cacheControl: cacheControl,
	}
}

func (l *localFileSystem) Exists(prefix string, filepath string) bool {
	for _, v := range notStaticPrefixes {
		if strings.HasPrefix(filepath, v) {
			return false
		}
	}
	//nolint:nestif
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)
		stats, err := os.Stat(name)
		if err != nil {
			return false
		}
		if stats.IsDir() {
			if !l.indexes {
				index := path.Join(name, INDEX)
				_, err := os.Stat(index)
				if err != nil {
					return false
				}
			}
		}
		return true
	}
	return false
}

func (l *localFileSystem) CacheControlHeader() string {
	return l.cacheControl
}

func StaticServeRedirect(urlPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		for _, v := range notStaticPrefixes {
			if strings.HasPrefix(requestPath, v) {
				return
			}
		}
		requestPath = strings.TrimPrefix(requestPath, urlPrefix)
		redirect := "/?next=" + requestPath + "?" + c.Request.URL.RawQuery
		c.Redirect(http.StatusFound, redirect)
		c.Abort()
	}
}
