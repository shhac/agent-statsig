package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	SetConfigDir(dir)
	t.Cleanup(func() { SetConfigDir("") })
}

func TestReadEmpty(t *testing.T) {
	setupTestDir(t)
	cfg := Read()
	if cfg == nil {
		t.Fatal("Read() returned nil")
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(cfg.Projects))
	}
}

func TestWriteAndRead(t *testing.T) {
	setupTestDir(t)
	cfg := &Config{
		DefaultProject: "prod",
		Projects: map[string]Project{
			"prod": {},
		},
	}
	if err := Write(cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	ClearCache()
	got := Read()
	if got.DefaultProject != "prod" {
		t.Errorf("DefaultProject = %q, want prod", got.DefaultProject)
	}
	if _, ok := got.Projects["prod"]; !ok {
		t.Error("project 'prod' not found after read")
	}
}

func TestStoreProject(t *testing.T) {
	setupTestDir(t)
	if err := StoreProject("staging", Project{}); err != nil {
		t.Fatalf("StoreProject: %v", err)
	}

	ClearCache()
	cfg := Read()
	if cfg.DefaultProject != "staging" {
		t.Errorf("first project should become default, got %q", cfg.DefaultProject)
	}

	if err := StoreProject("prod", Project{}); err != nil {
		t.Fatalf("StoreProject: %v", err)
	}
	ClearCache()
	cfg = Read()
	if cfg.DefaultProject != "staging" {
		t.Errorf("second project should not change default, got %q", cfg.DefaultProject)
	}
}

func TestRemoveProject(t *testing.T) {
	setupTestDir(t)
	StoreProject("a", Project{})
	StoreProject("b", Project{})

	ClearCache()
	cfg := Read()
	if cfg.DefaultProject != "a" {
		t.Fatalf("default should be 'a', got %q", cfg.DefaultProject)
	}

	RemoveProject("a")
	ClearCache()
	cfg = Read()
	if _, ok := cfg.Projects["a"]; ok {
		t.Error("project 'a' should be removed")
	}
	if cfg.DefaultProject == "a" {
		t.Error("default should no longer be 'a'")
	}
}

func TestSetDefault(t *testing.T) {
	setupTestDir(t)
	StoreProject("a", Project{})
	StoreProject("b", Project{})

	SetDefault("b")
	ClearCache()
	cfg := Read()
	if cfg.DefaultProject != "b" {
		t.Errorf("default = %q, want b", cfg.DefaultProject)
	}
}

func TestConfigDir(t *testing.T) {
	SetConfigDir("")
	defer SetConfigDir("")

	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-test")
	dir := ConfigDir()
	if dir != "/tmp/xdg-test/agent-statsig" {
		t.Errorf("ConfigDir() = %q", dir)
	}
}

func TestConfigFilePerms(t *testing.T) {
	setupTestDir(t)
	Write(&Config{Projects: make(map[string]Project)})

	info, err := os.Stat(filepath.Join(ConfigDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0o644 {
		t.Errorf("config file perms = %o, want 644", perm)
	}
}
