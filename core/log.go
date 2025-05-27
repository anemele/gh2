package core

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

var logger *slog.Logger

func init() {
	filename := fmt.Sprintf("log_%s.txt", time.Now().Format("2006-01-02"))
	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	logger = slog.New(slog.NewTextHandler(
		file,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		},
	))
}
