package config

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func InitLogger(level slog.Level) {
	filename := logFilePath
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		panic(err)
	}
	handler := slog.NewTextHandler(
		logFile,
		&slog.HandlerOptions{
			Level: level,
		},
	)
	logger = slog.New(handler)
}

func GetLogger() *slog.Logger {
	return logger
}
