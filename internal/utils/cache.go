package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func CacheFile(namespace, name string, data []byte) (cachedPath string, err error) {
	if !fs.ValidPath(namespace) || namespace == "." || strings.ContainsAny(namespace, "\\:") {
		return "", fmt.Errorf("invalid cache namespace %q", namespace)
	}

	if !fs.ValidPath(name) || strings.ContainsAny(name, "\\:") {
		return "", fmt.Errorf("invalid cache file name %q", name)
	}

	filename := path.Base(name)
	if filename == "." || filename == ".." || filepath.VolumeName(filename) != "" {
		return "", fmt.Errorf("invalid cache file name %q", name)
	}

	expectedHash := sha256.Sum256(data)

	cacheBase, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("get user cache directory: %w", err)
	}

	if cacheBase == "" {
		return "", errors.New("user cache directory is empty")
	}

	cacheBase, err = filepath.Abs(cacheBase)
	if err != nil {
		return "", fmt.Errorf("resolve user cache directory: %w", err)
	}

	cacheBase = filepath.Clean(cacheBase)

	for current := cacheBase; ; current = filepath.Dir(current) {
		info, statErr := os.Lstat(current)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return "", fmt.Errorf("refuse insecure cache ancestor %s: not a real directory", current)
			}

			if runtime.GOOS != "windows" && info.Mode().Perm()&0o022 != 0 {
				return "", fmt.Errorf("refuse insecure cache ancestor %s: writable by group or others", current)
			}
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			return "", fmt.Errorf("inspect cache ancestor %s: %w", current, statErr)
		}

		if filepath.Dir(current) == current {
			break
		}
	}

	cacheRoot := filepath.Join(cacheBase, filepath.FromSlash(namespace))
	if err = os.MkdirAll(cacheRoot, 0o700); err != nil {
		return "", fmt.Errorf("create private cache directory %s: %w", cacheRoot, err)
	}

	if err = validateCachePath(cacheRoot, cacheBase); err != nil {
		return "", fmt.Errorf("refuse insecure cache directory %s: %w", cacheRoot, err)
	}

	identityHash := sha256.New()
	_, _ = io.WriteString(identityHash, filename)
	_, _ = identityHash.Write([]byte{0})
	_, _ = identityHash.Write(expectedHash[:])
	hashDir := filepath.Join(cacheRoot, hex.EncodeToString(identityHash.Sum(nil)))
	filePath := filepath.Join(hashDir, filename)

	if cachedPath, exists, err := verifyCachedFile(hashDir, cacheBase, filePath, int64(len(data)), expectedHash); err != nil {
		return "", err
	} else if exists {
		return cachedPath, nil
	}

	tempDir, err := os.MkdirTemp(cacheRoot, ".install-")
	if err != nil {
		return "", fmt.Errorf("stage file in private cache %s: %w", cacheRoot, err)
	}

	tempOwned := true
	defer func() {
		if tempOwned {
			if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
				err = errors.Join(err, fmt.Errorf("remove private cache staging directory: %w", cleanupErr))
			}
		}
	}()

	if err = validateCachePath(tempDir, cacheBase); err != nil {
		return "", fmt.Errorf("refuse insecure cache staging directory: %w", err)
	}

	stageRoot, err := os.OpenRoot(tempDir)
	if err != nil {
		return "", fmt.Errorf("open private cache staging directory: %w", err)
	}

	file, fileErr := stageRoot.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if fileErr != nil {
		return "", errors.Join(fmt.Errorf("create staged cache file: %w", fileErr), stageRoot.Close())
	}

	writeErr := file.Chmod(0o600)
	if writeErr == nil {
		var written int

		written, writeErr = file.Write(data)
		if writeErr == nil && written != len(data) {
			writeErr = io.ErrShortWrite
		}
	}

	if writeErr == nil {
		writeErr = file.Sync()
	}

	if closeErr := file.Close(); closeErr != nil {
		writeErr = errors.Join(writeErr, fmt.Errorf("close staged cache file: %w", closeErr))
	}

	if rootCloseErr := stageRoot.Close(); rootCloseErr != nil {
		writeErr = errors.Join(writeErr, fmt.Errorf("close cache staging directory: %w", rootCloseErr))
	}

	if writeErr != nil {
		return "", fmt.Errorf("write staged cache file: %w", writeErr)
	}

	if err = verifyPrivateFile(filepath.Join(tempDir, filename), int64(len(data)), expectedHash); err != nil {
		return "", fmt.Errorf("staged cache file failed size or SHA-256 verification: %w", err)
	}

	if err = os.Rename(tempDir, hashDir); err != nil {
		if cachedPath, exists, cacheErr := verifyCachedFile(hashDir, cacheBase, filePath, int64(len(data)), expectedHash); cacheErr == nil && exists {
			return cachedPath, nil
		} else if cacheErr != nil {
			err = errors.Join(fmt.Errorf("atomically install cached file: %w", err), cacheErr)
		} else {
			err = fmt.Errorf("atomically install cached file: %w", err)
		}

		return "", err
	}

	tempOwned = false

	return filePath, nil
}

func validateCachePath(directory, cacheBase string) error {
	for current := filepath.Clean(directory); ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("%s is not a real directory", current)
		}

		if runtime.GOOS != "windows" {
			mode := info.Mode().Perm()
			relative, relErr := filepath.Rel(cacheBase, current)

			insideCache := relErr == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
			if insideCache && mode&0o077 != 0 {
				return fmt.Errorf("%s has non-private permissions %04o", current, mode)
			}

			if !insideCache && mode&0o022 != 0 {
				return fmt.Errorf("%s is writable by group or others", current)
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			return nil
		}
	}
}

func verifyCachedFile(hashDir, cacheBase, filePath string, size int64, expectedHash [sha256.Size]byte) (string, bool, error) {
	info, err := os.Lstat(hashDir)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("inspect cache directory: %w", err)
	}

	if err = validateCachePath(hashDir, cacheBase); err != nil {
		return "", false, fmt.Errorf("refuse insecure cache directory: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return "", false, fmt.Errorf("refuse non-directory cache entry %s", hashDir)
	}

	if err = verifyPrivateFile(filePath, size, expectedHash); err != nil {
		return "", false, fmt.Errorf("refuse invalid cached file %s: %w", filePath, err)
	}

	return filePath, true, nil
}

func verifyPrivateFile(filePath string, size int64, expectedHash [sha256.Size]byte) error {
	info, err := os.Lstat(filePath)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", filePath)
	}

	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%s has non-private permissions %04o", filePath, info.Mode().Perm())
	}

	if info.Size() != size {
		return fmt.Errorf("%s has unexpected size %d (want %d)", filePath, info.Size(), size)
	}

	root, err := os.OpenRoot(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	file, err := root.Open(filepath.Base(filePath))
	if err != nil {
		return errors.Join(err, root.Close())
	}

	hash := sha256.New()
	read, readErr := io.Copy(hash, file)
	closeErr := file.Close()
	rootCloseErr := root.Close()

	if readErr != nil || closeErr != nil || rootCloseErr != nil {
		return errors.Join(readErr, closeErr, rootCloseErr)
	}

	if read != size || !bytes.Equal(hash.Sum(nil), expectedHash[:]) {
		return fmt.Errorf("%s failed size or SHA-256 verification", filePath)
	}

	return nil
}
