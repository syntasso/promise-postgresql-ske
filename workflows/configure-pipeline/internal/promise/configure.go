package promise

import (
	"os"
	"time"

	"github.com/syntasso/promise-postgresql/postgresql-configure-pipeline/internal/logger"
)

func Configure() {
	const (
		sourceDir      = "/resources/dependencies"
		destinationDir = "/kratix/output"
	)

	logger := logger.New("promise-configure")
	start := time.Now()

	logger.Info("PostgreSQL promise configure started")

	fileCount, err := CopyTree(sourceDir, destinationDir, func(sourcePath, destinationPath string) {
		logger.Info("copying dependency file",
			"source", sourcePath,
			"destination", destinationPath,
		)
	})
	if err != nil {
		logger.Error(err, "PostgreSQL promise configure failed",
			"source", sourceDir,
			"destination", destinationDir,
		)
		os.Exit(1)
	}

	logger.Info("PostgreSQL promise configure completed",
		"file_count", fileCount,
		"duration_ms", time.Since(start).Milliseconds(),
		"result", "dependencies published",
	)
}
