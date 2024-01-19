package partupload

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

func getSHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

func saveToFile(src io.ReadCloser, targetPath string) error {
	lockPath := targetPath + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to lock the chunk file: %w", err)
	}

	defer func() {
		_ = lock.Close()
		_ = os.Remove(lockPath)
	}()

	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}

	defer func() {
		_ = target.Close()
	}()

	buf := make([]byte, 1024)

	for {
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read the chunk data: %w", err)
		}

		if n == 0 {
			break
		}

		if _, err := target.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write the chunk data: %w", err)
		}
	}

	return nil
}

func isFileLocked(path string) bool {
	lockPath := path + ".lock"

	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return false
	}

	return true
}

func joinParts(targetPath string, parts ...string) error {
	target, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}

	defer func() {
		_ = target.Close()
	}()

	opened := make([]*os.File, 0, len(parts))
	defer func() {
		for _, src := range opened {
			_ = src.Close()
		}
	}()

	for _, part := range parts {
		src, err := os.Open(part)
		if err != nil {
			return fmt.Errorf("failed to open chunk file: %w", err)
		}

		opened = append(opened, src)

		buf := make([]byte, 1024)

		for {
			n, err := src.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("failed to read the chunk data: %w", err)
			}

			if n == 0 {
				break
			}

			if _, err := target.Write(buf[:n]); err != nil {
				return fmt.Errorf("failed to write the chunk data: %w", err)
			}
		}
	}

	return nil
}

func cleanup(paths ...string) error {
	var err error

	for _, path := range paths {
		errors.Join(err, os.Remove(path))
	}

	return err
}
