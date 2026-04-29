package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mirainya/nexus/pkg/config"
)

type Storage interface {
	Save(filename string, reader io.Reader) (path string, err error)
	Get(path string) (io.ReadCloser, error)
	Delete(path string) error
}

type localStorage struct {
	basePath string
}

func New() (Storage, error) {
	cfg := config.C.Storage
	switch cfg.Type {
	case "local", "":
		if err := os.MkdirAll(cfg.LocalPath, 0755); err != nil {
			return nil, err
		}
		return &localStorage{basePath: cfg.LocalPath}, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}

func (s *localStorage) Save(filename string, reader io.Reader) (string, error) {
	ext := filepath.Ext(filename)
	path := filepath.Join(s.basePath, uuid.New().String()+ext)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return "", err
	}
	return path, nil
}

func (s *localStorage) validatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	absBase, err := filepath.Abs(s.basePath)
	if err != nil {
		return fmt.Errorf("invalid base path: %w", err)
	}
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) && absPath != absBase {
		return fmt.Errorf("path traversal denied")
	}
	return nil
}

func (s *localStorage) Get(path string) (io.ReadCloser, error) {
	if err := s.validatePath(path); err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *localStorage) Delete(path string) error {
	if err := s.validatePath(path); err != nil {
		return err
	}
	return os.Remove(path)
}
