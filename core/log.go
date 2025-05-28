package core

import (
	"log/slog"
	"os"
)

var (
	logger *slog.Logger
)

func InitLogger(level slog.Level) {
	filename := "gh2.log"
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
