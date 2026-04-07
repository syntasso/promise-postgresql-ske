package promise

import (
	"os"
	"time"

	"github.com/syntasso/promise-postgresql/postgresql-configure-pipeline/internal/logger"
)

func copyDir(src, dst string) error {
	return os.CopyFS(dst, os.DirFS(src))
}

func Configure() {
	const (
		sourceDir      = "/resources/dependencies"
		destinationDir = "/kratix/output"
	)

	log := logger.New("promise-configure")
	start := time.Now()

	log.Info("PostgreSQL promise configure started")

	err := copyDir(sourceDir, destinationDir)
	if err != nil {
		log.Error(err, "PostgreSQL promise configure failed")
		os.Exit(1)
	}
	log.Info("files copied successfully",
		"sourceDir", sourceDir,
		"destinationDir", destinationDir)

	log.Info("PostgreSQL promise configure completed",
		"duration_ms", time.Since(start).Milliseconds(),
		"result", "dependencies published",
	)
}
