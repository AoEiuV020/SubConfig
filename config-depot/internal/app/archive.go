package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractConfigBundle(bundle []byte, targetDir string) error {
	parent := filepath.Dir(targetDir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("create config parent: %w", err)
	}

	tempDir, err := os.MkdirTemp(parent, ".config-depot-*")
	if err != nil {
		return fmt.Errorf("create temporary config dir: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(tempDir)
		}
	}()

	if err := extractTarGzip(bundle, tempDir); err != nil {
		return err
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("replace config dir: %w", err)
	}
	if err := os.Rename(tempDir, targetDir); err != nil {
		return fmt.Errorf("commit config dir: %w", err)
	}
	committed = true
	return nil
}

func extractTarGzip(bundle []byte, targetDir string) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(bundle))
	if err != nil {
		return fmt.Errorf("read gzip bundle: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		destination, err := safeArchivePath(targetDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return fmt.Errorf("create archive dir: %w", err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
				return fmt.Errorf("create archive file parent: %w", err)
			}
			file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
			if err != nil {
				return fmt.Errorf("create archive file: %w", err)
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				_ = file.Close()
				return fmt.Errorf("write archive file: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("close archive file: %w", err)
			}
		default:
			return fmt.Errorf("unsupported archive entry type %d for %s", header.Typeflag, header.Name)
		}
	}
}

func safeArchivePath(targetDir string, name string) (string, error) {
	if name == "" {
		return "", errors.New("archive entry name is empty")
	}
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("archive entry must be relative: %s", name)
	}

	cleanName := filepath.Clean(name)
	if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry escapes config dir: %s", name)
	}

	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	destination := filepath.Join(targetDir, cleanName)
	if destination != targetDir && !strings.HasPrefix(destination, targetDir+string(filepath.Separator)) {
		return "", fmt.Errorf("archive entry escapes config dir: %s", name)
	}
	return destination, nil
}
