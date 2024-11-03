package log

import (
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
)

var (
	once   sync.Once
	logger *slog.Logger
	file   *os.File
)

// InitializeLogger will be called at the main.go once and create the logger according to the environment,
// sets the logger in the slog.Logger to ensue same logger will be reused if logger is required.
func InitializeLogger(env constants.Environment, serviceName string) {
	// Initialize logger configuration based on environment
	once.Do(func() {
		switch env {
		case constants.Local:
			file, err := os.OpenFile(serviceName+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				log.Panic(err)
			}
			handler := slog.NewJSONHandler(file, &slog.HandlerOptions{
				AddSource: true,
			})
			logger = slog.New(handler)
		default:
			logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				AddSource: true,
			}))
		}

		slog.SetDefault(logger)
	})
}

func Close() {
	if file != nil {
		file.Close()
	}
}
