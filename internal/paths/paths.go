package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	dirName = "curlew"

	dirPerm = 0o700
)

type Dirs struct {
	Config string

	Cache string
}

func Resolve() (Dirs, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return Dirs{}, fmt.Errorf("locate user config dir: %w", err)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return Dirs{}, fmt.Errorf("locate user cache dir: %w", err)
	}
	return Dirs{
		Config: filepath.Join(cfg, dirName),
		Cache:  filepath.Join(cache, dirName),
	}, nil
}

func (d Dirs) Collections() string { return filepath.Join(d.Config, "collections") }

func (d Dirs) DB() string { return filepath.Join(d.Config, "curlew.db") }

func (d Dirs) HistoryDB() string { return filepath.Join(d.Cache, "history.db") }

func (d Dirs) Blobs() string { return filepath.Join(d.Cache, "blobs") }

func (d Dirs) MkdirAll() error {
	for _, dir := range []string{d.Config, d.Cache, d.Collections(), d.Blobs()} {
		if err := os.MkdirAll(dir, dirPerm); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
