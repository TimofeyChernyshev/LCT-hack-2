package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) *LocalStorage { return &LocalStorage{root: root} }

func (s *LocalStorage) Save(_ context.Context, userID, filename string, r io.Reader, maxBytes int64) (string, error) {
	dir := filepath.Join(s.root, userID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	name := uuid.NewString() + filepath.Ext(filename)
	full := filepath.Join(dir, name)

	f, err := os.Create(full)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, io.LimitReader(r, maxBytes+1)); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return full, nil
}
