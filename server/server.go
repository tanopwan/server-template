package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	GCP     = "GCP"
	CONSOLE = "Console"
)

type Instance struct {
	http.Server
	loggerType string
	appName    string
	hook       *GCPHook
}

var ContextKeyRequestID = "request_id"
var LogFieldRequestID = "request_id"
var LogFieldAppName = "app"

func NewInstance(appName string, appVersion string, router http.Handler) *Instance {
	port := getEnvOrDefault("PORT", "8080")
	loggerType := getEnvOrDefault("LOGGER_TYPE", CONSOLE)
	instance := Instance{
		Server: http.Server{
			Addr:           ":" + port,
			Handler:        router,
			ReadTimeout:    20 * time.Second,
			WriteTimeout:   20 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		appName: appName,
	}

	projectID, ok := os.LookupEnv("PROJECT_ID")
	if ok && loggerType == GCP {
		// Hooking StackDriver log
		hook, err := NewGCPHook(context.Background(), projectID, appName)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.AddHook(hook)
		logrus.Info("Add hook to GCP")
	}
	logrus.Infof("%s [%s] version %s at %s", appName, projectID, appVersion, port)
	return &instance
}

func (i *Instance) Start() error {
	go func() {
		err := i.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("[ERR] server exited with: %s", err)
		}
	}()

	stop := make(chan os.Signal, 1)

	// pkill -15 main
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	logrus.Info("Start shutting down server")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if i.hook != nil {
		err := i.hook.Close()
		if err != nil {
			logrus.Errorf("Failed to close logging client with error: %+v", err)
		} else {
			logrus.Info("Successfully closed logging client")
		}
	}
	err := i.Shutdown(ctxShutDown)
	if err != nil {
		return err
	}
	logrus.Info("Server exit properly")
	return nil
}

func (i *Instance) GetFieldLoggerFromCtx(ctx context.Context) logrus.FieldLogger {
	reqID := ctx.Value(ContextKeyRequestID)

	var requestID string
	if value, ok := reqID.(string); ok {
		requestID = value
	}

	if requestID == "" {
		requestID, _ = generateRequestID()
	}

	return logrus.WithFields(logrus.Fields{LogFieldRequestID: requestID})
}

func (i *Instance) GetLoggerFromCtx(ctx context.Context) logrus.StdLogger {
	reqID := ctx.Value(ContextKeyRequestID)

	var requestID string
	if value, ok := reqID.(string); ok {
		requestID = value
	}

	if requestID == "" {
		requestID, _ = generateRequestID()
	}

	return logrus.WithFields(logrus.Fields{LogFieldRequestID: requestID, LogFieldAppName: i.appName})
}

func getEnvOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func generateRequestID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to rand.Read with reason: %w", err)
	}

	return hex.EncodeToString(b), nil
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
