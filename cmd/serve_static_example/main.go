package main

import (
	"log"

	"github.com/tanopwan/server-template/server"
)

// go run cmd/serve_static_example/main.go
func main() {
	srv := server.NewInstance("example-app", "1", nil)
	srv.DefaultServeMuxServeFilesWithCacheControl(server.WebServerConfig{
		PublicDir:          "cmd/serve_static_example/public",
		CacheControlMaxAge: 3600,
		ServeIndexFilePath: "cmd/serve_static_example/public/index.html",
		ServeStaticPath:    "/static/",
	})
	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}
