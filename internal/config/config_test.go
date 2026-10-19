package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		want    []string
	}{
		{
			name:    "valid file",
			content: "repos:\n  - a7d-corp/github-pr-summarizer\n  - giantswarm/happa\n",
			want:    []string{"a7d-corp/github-pr-summarizer", "giantswarm/happa"},
		},
		{
			name:    "malformed yaml",
			content: "repos: [\n",
			wantErr: true,
		},
		{
			name:    "empty repos",
			content: "repos: []\n",
			wantErr: true,
		},
		{
			name:    "missing repos key",
			content: "foo: bar\n",
			wantErr: true,
		},
		{
			name:    "malformed owner/name entry",
			content: "repos:\n  - not-a-valid-entry\n",
			wantErr: true,
		},
		{
			name:    "entry with too many slashes",
			content: "repos:\n  - a/b/c\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempConfig(t, tt.content)
			cfg, err := Load(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if len(cfg.Repos) != len(tt.want) {
				t.Fatalf("Load() repos = %v, want %v", cfg.Repos, tt.want)
			}
			for i, r := range tt.want {
				if cfg.Repos[i] != r {
					t.Errorf("Load() repos[%d] = %q, want %q", i, cfg.Repos[i], r)
				}
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
}

func TestAllowlist(t *testing.T) {
	cfg := Config{Repos: []string{"a7d-corp/github-pr-summarizer", "giantswarm/happa"}}
	allow := cfg.Allowlist()

	if _, ok := allow["a7d-corp/github-pr-summarizer"]; !ok {
		t.Error("Allowlist() missing a7d-corp/github-pr-summarizer")
	}
	if _, ok := allow["giantswarm/happa"]; !ok {
		t.Error("Allowlist() missing giantswarm/happa")
	}
	if _, ok := allow["not/present"]; ok {
		t.Error("Allowlist() unexpectedly contains not/present")
	}
	if len(allow) != 2 {
		t.Errorf("Allowlist() len = %d, want 2", len(allow))
	}
}

func TestDefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/home/example/.config")

	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() unexpected error: %v", err)
	}

	want := "/home/example/.config/github-pr-summarizer/repos.yaml"
	if path != want {
		t.Errorf("DefaultPath() = %q, want %q", path, want)
	}
}
