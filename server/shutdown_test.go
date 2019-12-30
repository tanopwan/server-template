package server_test

import (
	"github.com/julienschmidt/httprouter"
	"github.com/tanopwan/server-template/server"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestShutDownWithNormalHandler(t *testing.T) {
	t.Logf("Server starting")
	testServer := server.NewInstance("test-app", "1", nil)
	http.HandleFunc("/delay", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`Hello`))
		time.Sleep(10 * time.Second)
		writer.Write([]byte(`World`))
	})

	go callGetDelay10Endpoint(t)
	go sendSignalToServer(t)

	testServer.Start()
	t.Logf("Server stopped!")
}

func TestShutDownWithNormalHandlerAndLoggingClient(t *testing.T) {
	t.Logf("Server starting")
	os.Setenv("PROJECT_ID", "project_id")
	testServer := server.NewInstance("test-app", "1", nil)
	http.HandleFunc("/delay", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`Hello`))
		time.Sleep(10 * time.Second)
		writer.Write([]byte(`World`))
	})

	go callGetDelay10Endpoint(t)
	go sendSignalToServer(t)

	testServer.Start()
	t.Logf("Server stopped!")
}

func TestShutDownWithHttpRouter(t *testing.T) {
	t.Logf("Server starting")
	handler := httprouter.New()
	handler.GET("/delay", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Write([]byte(`Hello, `))
		time.Sleep(10 * time.Second)
		writer.Write([]byte(`World`))
	})
	testServer := server.NewInstance("test-app", "1", handler)

	go callGetDelay10Endpoint(t)
	go sendSignalToServer(t)

	testServer.Start()
	t.Logf("Server stopped!")
}

func TestShutDownWithNormalHandlerTimeout(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	t.Logf("Server starting")
	testServer := server.NewInstance("test-app", "1", nil)
	http.HandleFunc("/delay", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(`Hello`))
		time.Sleep(15 * time.Second)
		writer.Write([]byte(`World`))
	})

	go callGetDelay10Endpoint(t)
	go sendSignalToServer(t)

	testServer.Start()
	t.Logf("Server stopped!")
}

func TestShutDownWithHttpRouterTimeout(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	t.Logf("Server starting")
	handler := httprouter.New()
	handler.GET("/delay", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.Write([]byte(`Hello, `))
		time.Sleep(15 * time.Second)
		writer.Write([]byte(`World`))
	})
	testServer := server.NewInstance("test-app", "1", handler)

	go callGetDelay10Endpoint(t)
	go sendSignalToServer(t)

	testServer.Start()
	t.Logf("Server stopped!")
}

func callGetDelay10Endpoint(t *testing.T) {
	time.Sleep(time.Second)
	t.Logf("Wait one seconds and then call GET /delay")
	resp, err := http.Get("http://localhost:8080/delay")
	if err != nil {
		t.Errorf("failed to http GET with err: %s", err.Error())
		t.FailNow()
	}
	defer resp.Body.Close()
	t.Logf("Controller return successfully")
	bb, err := ioutil.ReadAll(resp.Body)
	t.Logf("Read %s", bb)
}

func sendSignalToServer(t *testing.T) {
	time.Sleep(3 * time.Second)
	t.Logf("Wait three seconds and then send sigterm")
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	t.Logf("Finish sent sigterm")
}
