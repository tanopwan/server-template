package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/tanopwan/server-template/server"
)

func main() {
	router := httprouter.New()
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "Welcome!\n")
	})
	err := server.NewInstance("example-app", "1", router).Start()
	if err != nil {
		os.Exit(1)
	}
}
