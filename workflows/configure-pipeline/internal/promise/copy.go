package promise

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CopyTree(sourceDir, destinationDir string, onFile func(sourcePath, destinationPath string)) (int, error) {
	fileCount := 0
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relativePath == "." {
			return os.MkdirAll(destinationDir, 0o755)
		}

		destinationPath := filepath.Join(destinationDir, relativePath)
		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(destinationPath, info.Mode())
		}

		if onFile != nil {
			onFile(path, destinationPath)
		}

		if err := copyFile(path, destinationPath, d); err != nil {
			return err
		}

		fileCount++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("copy tree %s to %s: %w", sourceDir, destinationDir, err)
	}

	return fileCount, nil
}

func copyFile(sourcePath, destinationPath string, entry fs.DirEntry) error {
	info, err := entry.Info()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		destinationFile.Close()
		return err
	}

	if err := destinationFile.Close(); err != nil {
		return err
	}

	return os.Chmod(destinationPath, info.Mode())
}
