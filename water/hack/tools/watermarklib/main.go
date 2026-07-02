// This command prepares the architecture-specific watermark static library for
// the external cgo package without modifying the package source files.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	libraryDir  = "backend/internal/library/watermark"
	outputName  = "libwatermark.a"
	fileMode    = 0o644
	dirFileMode = 0o755
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	arch := strings.TrimSpace(os.Getenv("GOARCH"))
	if arch == "" {
		arch = runtime.GOARCH
	}
	sourceName := "lib" + arch + "_watermark.a"
	sourcePath := filepath.Join(libraryDir, sourceName)
	targetPath := filepath.Join(libraryDir, outputName)
	if err := copyFile(sourcePath, targetPath); err != nil {
		return fmt.Errorf("prepare watermark static library for %s: %w", arch, err)
	}
	fmt.Printf("prepared %s from %s\n", targetPath, sourceName)
	return nil
}

func copyFile(sourcePath string, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	if err = os.MkdirAll(filepath.Dir(targetPath), dirFileMode); err != nil {
		return err
	}
	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	if _, err = io.Copy(target, source); err != nil {
		_ = target.Close()
		return err
	}
	return target.Close()
}
