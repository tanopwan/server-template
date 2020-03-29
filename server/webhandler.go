package server

import (
	"net/http"
	"strings"
)

type WebServerConfig struct {
	PublicDir          string
	CacheControlMaxAge int
	ServeIndexFilePath string
	ServeStaticPath    string
}

func (i *Instance) DefaultServeMuxServeFilesWithCacheControl(config WebServerConfig) {
	contextPath := config.ServeStaticPath
	if contextPath == "" {
		contextPath = "/"
	}
	if !strings.HasSuffix(contextPath, "/") {
		contextPath = contextPath + "/"
	}
	if !strings.HasPrefix(contextPath, "/") {
		contextPath = "/" + contextPath
	}

	fs := http.FileServer(http.Dir(config.PublicDir))
	http.DefaultServeMux.HandleFunc(contextPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age="+string(config.CacheControlMaxAge))
		fs.ServeHTTP(w, r)
	})

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private; max-age=0;")
		http.ServeFile(w, r, config.ServeIndexFilePath)
	})
}
