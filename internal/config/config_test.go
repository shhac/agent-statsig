package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

func TestSetDefaultNonexistent(t *testing.T) {
	setupTestDir(t)
	StoreProject("a", Project{})

	SetDefault("nonexistent")
	ClearCache()
	cfg := Read()
	if cfg.DefaultProject != "a" {
		t.Errorf("default should remain 'a' when setting nonexistent, got %q", cfg.DefaultProject)
	}
}

// Concurrent StoreProject calls must not lose each other's entries.
//
// Before updateConfig routed through creds.Store.Update, StoreProject did
// Read() (from the shared in-memory cache) -> mutate -> Write(). Two
// concurrent CLI invocations (in-process, sharing the package cache, or
// across processes sharing config.json) each built their write from a
// snapshot taken before the other's landed, so all but the last writer's
// project were silently erased.
func TestConcurrentStoreProjectDoesNotLoseEntries(t *testing.T) {
	setupTestDir(t)

	const writers = 20
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := StoreProject(fmt.Sprintf("project-%02d", i), Project{}); err != nil {
				t.Errorf("StoreProject: %v", err)
			}
		}(i)
	}
	wg.Wait()

	ClearCache()
	cfg := Read()
	if len(cfg.Projects) != writers {
		t.Fatalf("%d of %d concurrent StoreProject calls survived — updates were lost", len(cfg.Projects), writers)
	}
	for i := range writers {
		name := fmt.Sprintf("project-%02d", i)
		if _, ok := cfg.Projects[name]; !ok {
			t.Errorf("%s was lost from config.json", name)
		}
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

// config.json now goes through creds.Store (see updateConfig), which writes
// every file 0600 regardless of content — one audited place to get file
// permissions right rather than a per-file policy. That's a tightening from
// the previous 0644 default, not a regression: this directory is shared with
// credentials.json, which can hold a real API key when the keychain is
// unavailable.
func TestConfigFilePerms(t *testing.T) {
	setupTestDir(t)
	Write(&Config{Projects: make(map[string]Project)})

	info, err := os.Stat(filepath.Join(ConfigDir(), "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("config file perms = %o, want 600", perm)
	}
}
