package server

import (
	"cloud.google.com/go/logging"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type LoggerType string

const (
	GCP     LoggerType = "GCP"
	CONSOLE            = "Console"
	NONE               = "None"
)

type Instance struct {
	http.Server
	loggerType    LoggerType
	appName       string
	loggingClient *logging.Client
}

var ContextKeyRequestID = "request_id"
var LogFieldRequestID = "request_id"
var LogFieldAppName = "app"

var configLoggerType LoggerType = CONSOLE

func SetLogger(loggerType LoggerType) {
	configLoggerType = loggerType
}

func NewInstance(appName string, appVersion string, router http.Handler) *Instance {
	port := getEnvOrDefault("PORT", "8080")
	instance := Instance{
		Server: http.Server{
			Addr:           ":" + port,
			Handler:        router,
			ReadTimeout:    20 * time.Second,
			WriteTimeout:   20 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		loggerType: configLoggerType,
		appName:    appName,
	}

	projectID, ok := os.LookupEnv("PROJECT_ID")
	if ok && configLoggerType == GCP {
		// Use StackDriver logging
		ctx := context.Background()
		client, err := logging.NewClient(ctx, projectID)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		instance.loggingClient = client
		client.Logger(appName).StandardLogger(logging.Info).Printf("[GCP] %s [%s] version %s at %s", appName, projectID, appVersion, port)
	} else if configLoggerType == CONSOLE {
		logrus.WithField("app", appName).Infof("[Console] %s [%s] version %s at %s", appName, projectID, appVersion, port)
	}

	return &instance
}

func (i *Instance) Start() {

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

	log.Printf("Start shutting down server\n")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if i.loggingClient != nil {
		err := i.loggingClient.Close()
		if err != nil {
			log.Printf("Failed to close logging client with error: %s", err)
		} else {
			log.Printf("Successfully closed logging client")
		}
	}
	err := i.Shutdown(ctxShutDown)
	if err != nil {
		log.Panicf("Error shutting down server with error: %s", err)
	}
	log.Printf("Server exit properly\n")
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

	return logrus.WithFields(logrus.Fields{LogFieldRequestID: requestID, LogFieldAppName: i.appName})
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

	if i.loggerType == GCP {
		return i.loggingClient.Logger(i.appName).StandardLogger(logging.Info)
	} else if configLoggerType == CONSOLE {
		// OTHER Logger Types
		return logrus.WithFields(logrus.Fields{LogFieldRequestID: requestID, LogFieldAppName: i.appName})
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
