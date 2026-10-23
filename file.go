package caelum

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Rotation int

const (
	RotationNone Rotation = iota
	RotationRename
	RotationCompress
)

type FileConfig struct {
	Path     string
	Rotation Rotation
}

type FileWriter struct {
	file         *os.File
	compressDone <-chan struct{}
	compressErr  error

	mu        sync.Mutex
	closed    bool
	closeDone chan struct{}
	closeErr  error
}

func OpenFile(cfg FileConfig) (*FileWriter, error) {
	if cfg.Path == "" {
		return nil, errors.New("file path is empty")
	}

	if cfg.Rotation < RotationNone || cfg.Rotation > RotationCompress {
		return nil, fmt.Errorf("invalid file rotation: %d", cfg.Rotation)
	}

	path, err := filepath.Abs(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve log file %q: %w", cfg.Path, err)
	}

	cfg.Path = path
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o750); err != nil {
		return nil, fmt.Errorf("create file directory for %q: %w", cfg.Path, err)
	}

	var (
		rollback       func() error
		compressSource string
		compressPath   string
		compressFile   *os.File
	)

	if cfg.Rotation != RotationNone {
		exists, err := rotatableFileExists(cfg.Path)
		if err != nil {
			return nil, err
		}

		if exists {
			timestamp := time.Now().Format("2006-01-02_15-04-05.000000000")

			switch cfg.Rotation {
			case RotationNone:
			case RotationRename:
				archive, rotated, err := rotateRename(cfg.Path, timestamp)
				if err != nil {
					return nil, err
				}

				if rotated {
					rollback = func() error { return restoreRotatedFile(archive, cfg.Path) }
				}
			case RotationCompress:
				source, archive, output, rotated, err := rotateForCompression(cfg.Path, timestamp)
				if err != nil {
					return nil, err
				}

				if rotated {
					compressSource, compressPath, compressFile = source, archive, output
					rollback = func() error {
						closeErr := output.Close()
						removeErr := removeIfExists(archive)
						restoreErr := restoreRotatedFile(source, cfg.Path)

						return errors.Join(closeErr, removeErr, restoreErr)
					}
				}
			}
		}
	}

	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		if rollback != nil {
			return nil, errors.Join(fmt.Errorf("open log file %q: %w", cfg.Path, err), rollback())
		}

		return nil, fmt.Errorf("open log file %q: %w", cfg.Path, err)
	}

	writer := &FileWriter{file: file, closeDone: make(chan struct{})}
	if compressFile != nil {
		done := make(chan struct{})
		writer.compressDone = done

		go func() {
			err := compressRotatedFile(compressSource, compressPath, compressFile)

			writer.mu.Lock()
			writer.compressErr = err
			writer.mu.Unlock()
			close(done)
		}()
	}

	return writer, nil
}

func (w *FileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, os.ErrClosed
	}

	return w.file.Write(p)
}

func (w *FileWriter) Flush() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()

		return os.ErrClosed
	}

	err := w.file.Sync()
	done := w.compressDone
	w.mu.Unlock()

	if done != nil {
		<-done
	}

	w.mu.Lock()
	err = errors.Join(err, w.compressErr)
	w.mu.Unlock()

	return err
}

func (w *FileWriter) Unwrap() io.Writer {
	return w.file
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		done := w.closeDone
		w.mu.Unlock()
		<-done

		return w.closeErr
	}

	w.closed = true
	fileErr := errors.Join(w.file.Sync(), w.file.Close())
	compressDone := w.compressDone
	w.mu.Unlock()

	if compressDone != nil {
		<-compressDone
	}

	w.mu.Lock()
	closeErr := errors.Join(fileErr, w.compressErr)
	w.closeErr = closeErr
	close(w.closeDone)
	w.mu.Unlock()

	return closeErr
}

func rotatableFileExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("inspect log file %q: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("rotate log file %q: symbolic links are not supported", path)
	}

	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("rotate log file %q: not a regular file", path)
	}

	return true, nil
}

func rotatedPath(path, timestamp string, collision int, suffix string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)

	collisionSuffix := ""
	if collision > 0 {
		collisionSuffix = fmt.Sprintf("-%d", collision)
	}

	return filepath.Join(filepath.Dir(path), stem+"."+timestamp+collisionSuffix+ext+suffix)
}

func reservePath(path, timestamp, suffix string) (string, *os.File, error) {
	for collision := 0; ; collision++ {
		candidate := rotatedPath(path, timestamp, collision, suffix)

		file, err := os.OpenFile(candidate, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}

		if err != nil {
			return "", nil, fmt.Errorf("reserve rotated file %q: %w", candidate, err)
		}

		return candidate, file, nil
	}
}

func rotateRename(path, timestamp string) (string, bool, error) {
	archive, reservation, err := reservePath(path, timestamp, "")
	if err != nil {
		return "", false, err
	}

	if err := reservation.Close(); err != nil {
		return "", false, errors.Join(fmt.Errorf("close archive reservation %q: %w", archive, err), removeIfExists(archive))
	}

	rotated, err := renameReserved(path, archive)
	if err != nil {
		return "", false, fmt.Errorf("rotate log file %q: %w", path, err)
	}

	return archive, rotated, nil
}

func rotateForCompression(path, timestamp string) (string, string, *os.File, bool, error) {
	var (
		source  string
		archive string
		output  *os.File
	)

	for collision := 0; ; collision++ {
		source = rotatedPath(path, timestamp, collision, ".compressing")

		reservation, err := os.OpenFile(source, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}

		if err != nil {
			return "", "", nil, false, fmt.Errorf("reserve temporary rotated file %q: %w", source, err)
		}

		if err := reservation.Close(); err != nil {
			return "", "", nil, false, errors.Join(fmt.Errorf("close temporary reservation %q: %w", source, err), removeIfExists(source))
		}

		archive = rotatedPath(path, timestamp, collision, ".gz")
		output, err = os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)

		if errors.Is(err, os.ErrExist) {
			if removeErr := removeIfExists(source); removeErr != nil {
				return "", "", nil, false, errors.Join(fmt.Errorf("reserve compressed archive %q: %w", archive, err), removeErr)
			}

			continue
		}

		if err != nil {
			return "", "", nil, false, errors.Join(fmt.Errorf("reserve compressed archive %q: %w", archive, err), removeIfExists(source))
		}

		break
	}

	rotated, err := renameReserved(path, source)
	if err != nil {
		return "", "", nil, false, errors.Join(fmt.Errorf("rotate log file %q for compression: %w", path, err), output.Close(), removeIfExists(archive))
	}

	if !rotated {
		if err := errors.Join(output.Close(), removeIfExists(archive)); err != nil {
			return "", "", nil, false, err
		}

		return "", "", nil, false, nil
	}

	return source, archive, output, true, nil
}

func renameReserved(source, destination string) (bool, error) {
	if err := os.Rename(source, destination); err != nil {
		cleanupErr := removeIfExists(destination)
		if sourceNoLongerExists(source) {
			return false, cleanupErr
		}

		return false, errors.Join(err, cleanupErr)
	}

	return true, nil
}

func sourceNoLongerExists(path string) bool {
	_, err := os.Lstat(path)

	return errors.Is(err, os.ErrNotExist)
}

func restoreRotatedFile(archive, original string) error {
	if err := os.Link(archive, original); err != nil {
		return fmt.Errorf("restore rotated file %q to %q: %w", archive, original, err)
	}

	if err := os.Remove(archive); err != nil {
		return fmt.Errorf("remove restored archive %q: %w", archive, err)
	}

	return nil
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func compressRotatedFile(source, archive string, output *os.File) error {
	input, err := os.Open(source)
	if err != nil {
		return errors.Join(fmt.Errorf("open rotated file %q for compression: %w", source, err), output.Close(), removeIfExists(archive))
	}

	compressed := gzip.NewWriter(output)
	_, copyErr := io.Copy(compressed, input)
	inputCloseErr := input.Close()
	compressionCloseErr := compressed.Close()
	syncErr := output.Sync()

	outputCloseErr := output.Close()
	if err := errors.Join(copyErr, inputCloseErr, compressionCloseErr, syncErr, outputCloseErr); err != nil {
		return errors.Join(fmt.Errorf("compress rotated file %q: %w", source, err), removeIfExists(archive))
	}

	if err := os.Remove(source); err != nil {
		return fmt.Errorf("remove compressed source file %q: %w", source, err)
	}

	return nil
}
