package server

import (
	"context"
	"fmt"

	"cloud.google.com/go/logging"
	"github.com/sirupsen/logrus"
)

type GCPHook struct {
	appName       string
	projectID     string
	logger        *logging.Logger
	loggingClient *logging.Client
}

var logLevelMappings = map[logrus.Level]logging.Severity{
	logrus.TraceLevel: logging.Default,
	logrus.DebugLevel: logging.Debug,
	logrus.InfoLevel:  logging.Info,
	logrus.WarnLevel:  logging.Warning,
	logrus.ErrorLevel: logging.Error,
	logrus.FatalLevel: logging.Critical,
	logrus.PanicLevel: logging.Critical,
}

func NewGCPHook(ctx context.Context, projectID, appName string) (*GCPHook, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	hook := &GCPHook{
		appName:       appName,
		projectID:     projectID,
		loggingClient: client,
	}
	return hook, nil
}

func (h *GCPHook) Fire(entry *logrus.Entry) error {
	if h.logger == nil {
		h.logger = h.loggingClient.Logger(h.appName)
	}
	entry.Data["app"] = h.appName
	entry.Data["project_id"] = h.projectID
	payload, err := entry.String()
	if err != nil {
		return err
	}
	h.logger.Log(logging.Entry{Payload: payload, Severity: logLevelMappings[entry.Level]})
	return nil
}

func (h *GCPHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *GCPHook) Close() error {
	return h.loggingClient.Close()
}
