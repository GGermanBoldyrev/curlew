package paths_test

import (
	"path/filepath"
	"testing"

	"github.com/GGermanBoldyrev/curlew/internal/paths"
)

func TestDirsLayout(t *testing.T) {
	t.Parallel()

	d := paths.Dirs{Config: filepath.Join("cfg", "curlew"), Cache: filepath.Join("cache", "curlew")}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"collections under config", d.Collections(), filepath.Join("cfg", "curlew", "collections")},
		{"history under cache", d.HistoryDB(), filepath.Join("cache", "curlew", "history.db")},
		{"blobs under cache", d.Blobs(), filepath.Join("cache", "curlew", "blobs")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestMkdirAllCreatesEverything(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	d := paths.Dirs{Config: filepath.Join(root, "config"), Cache: filepath.Join(root, "cache")}

	if err := d.MkdirAll(); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for _, dir := range []string{d.Config, d.Cache, d.Collections(), d.Blobs()} {
		if _, err := filepath.Abs(dir); err != nil {
			t.Fatalf("abs %s: %v", dir, err)
		}
	}
}

func TestResolveIsAbsolute(t *testing.T) {
	t.Parallel()

	d, err := paths.Resolve()
	if err != nil {
		t.Skipf("no user directories in this environment: %v", err)
	}
	if !filepath.IsAbs(d.Config) || !filepath.IsAbs(d.Cache) {
		t.Errorf("expected absolute paths, got config=%q cache=%q", d.Config, d.Cache)
	}
}
