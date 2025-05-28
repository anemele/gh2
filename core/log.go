package core

import (
	"bufio"
	"log/slog"
	"os"
)

var (
	logger    *slog.Logger
	logFile   *os.File
	bufWriter *bufio.Writer
)

func InitLogger(level slog.Level) {
	filename := "gh2.log"
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	bufWriter = bufio.NewWriterSize(logFile, 64<<10) // 64KB
	handler := slog.NewTextHandler(
		bufWriter,
		&slog.HandlerOptions{
			Level: level,
		},
	)
	logger = slog.New(handler)
}

func GetLogger() *slog.Logger {
	return logger
}

func CloseLogger() error {
	if bufWriter != nil {
		if err := bufWriter.Flush(); err != nil {
			return err
		}
	}
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}
