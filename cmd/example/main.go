package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/tanopwan/server-template/auth"
	"github.com/tanopwan/server-template/server"
)

func run() error {
	firebaseAuthService := auth.NewFirebaseAuthService()
	router := httprouter.New()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world")
	})

	srv := server.NewInstance("example-app", "1", handler)
	router.HandlerFunc("POST", "/api/auth/firebase-auth/register", func(writer http.ResponseWriter, request *http.Request) {
		logger := srv.GetFieldLoggerFromCtx(request.Context())
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			logger.Errorf("handler err: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		info := auth.RegisterFirebaseAuthInfo{}
		err = json.Unmarshal(body, &info)
		if err != nil {
			logger.Errorf("handler err: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		u, err := firebaseAuthService.Register(request.Context(), info)
		if err != nil {
			logger.Errorf("handler err: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(writer, "failed to register: %v", err)
			return
		}

		bb, err := json.Marshal(u)
		if err != nil {
			logger.Errorf("handler err: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Infof("successfully created user UID = %s", u.UID)
		writer.Write(bb)
	})
	return srv.Start()
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}
