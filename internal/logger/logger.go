package logger

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lacolle87/eqmlog"
)

func SetupLogger(logFile string) {
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	multiWriter := eqmlog.LoadLogger()
	logger := slog.New(slog.NewTextHandler(multiWriter, nil))
	slog.SetDefault(logger)
	slog.Info("Logger initialized", "file", logFile)
}
