package server

import (
	"cloud.google.com/go/logging"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Instance struct {
	http.Server
	info  *log.Logger
	error *log.Logger
}

func NewInstance(appName string, appVersion string, router http.Handler) *Instance {
	port := getEnvOrDefault("PORT", "8080")
	instance := Instance{
		Server: http.Server{
			Addr:           ":" + port,
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	ctx := context.Background()
	projectID, ok := os.LookupEnv("PROJECT_ID")
	if ok {
		// Use StackDriver logging
		client, err := logging.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		// Sets the name of the log to write to.
		logName := strings.ReplaceAll(appName, " ", "")

		instance.info = client.Logger(logName).StandardLogger(logging.Info)
		instance.error = client.Logger(logName).StandardLogger(logging.Error)
		instance.info.Printf("[Enabled Log][Starting...] %s [%s] version %s at %s", appName, projectID, appVersion, port)
	}

	return &instance
}

func (i *Instance) Start() error {
	err := i.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to ListenAndServe with reason: %w", err)
	}

	return nil
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type ReverseProxyConfig struct {
	ProxyURL string
	AuthURL  string
	AuthPath string
}

func generateRequestID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to rand.Read with reason: %w", err)
	}

	return hex.EncodeToString(b), nil
}

func (i *Instance) InfoLog(r *http.Request, message string, format ...interface{}) {
	ctxRequestID := r.Context().Value("request_id")
	var requestID string
	if ctxRequestID == nil {
		requestID = ""
	} else {
		requestID = ctxRequestID.(string)
	}
	if i.info != nil {
		i.info.Printf("[%s] %s\n", requestID, fmt.Sprintf(message, format...))
	} else {
		fmt.Printf("[%s] %s\n", requestID, fmt.Sprintf(message, format...))
	}
}

func (i *Instance) ErrorLog(r *http.Request, message string, format ...interface{}) {
	ctxRequestID := r.Context().Value("request_id")
	var requestID string
	if ctxRequestID == nil {
		requestID = ""
	} else {
		requestID = ctxRequestID.(string)
	}
	if i.error != nil {
		i.error.Printf("[%s] %s\n", requestID, fmt.Sprintf(message, format...))
	} else {
		log.Printf("[%s] %s\n", requestID, fmt.Sprintf(message, format...))
	}
}

func AddRequestIDToContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		ctx := r.Context()
		if requestID == "" {
			value, err := generateRequestID()
			if err != nil {
				ctx = context.WithValue(ctx, "request_id", "fail-xxx")
				r = r.WithContext(ctx)
			}
			requestID = "req-" + value
		}
		ctx = context.WithValue(ctx, "request_id", requestID)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
