package project

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/shhac/agent-statsig/internal/config"
	"github.com/shhac/agent-statsig/internal/credential"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	config.SetConfigDir(dir)
	t.Cleanup(func() { config.SetConfigDir("") })
}

func TestRegisterCreatesAllSubcommands(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	Register(root)

	proj, _, err := root.Find([]string{"project"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"add", "update", "remove", "list", "set-default", "test"}
	for _, name := range expected {
		found := false
		for _, cmd := range proj.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestAddRequiresConsoleKey(t *testing.T) {
	setupTestDir(t)

	root := &cobra.Command{Use: "test"}
	Register(root)
	root.SetArgs([]string{"project", "add", "myproj"})

	// Should not error (writes to stderr instead)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	// No credential should be stored since --console-key was missing
	_, err := credential.Get("myproj")
	if err == nil {
		t.Error("credential should not exist without --console-key")
	}
}

func TestAddAndList(t *testing.T) {
	setupTestDir(t)

	root := &cobra.Command{Use: "test"}
	Register(root)

	root.SetArgs([]string{"project", "add", "testproj", "--console-key", "key123"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	cred, err := credential.Get("testproj")
	if err != nil {
		t.Fatal(err)
	}
	if cred.ConsoleKey != "key123" {
		t.Errorf("ConsoleKey = %q", cred.ConsoleKey)
	}

	cfg := config.Read()
	if cfg.DefaultProject != "testproj" {
		t.Errorf("default = %q, want testproj", cfg.DefaultProject)
	}
}

func TestRemoveNonexistent(t *testing.T) {
	setupTestDir(t)

	root := &cobra.Command{Use: "test"}
	Register(root)
	root.SetArgs([]string{"project", "remove", "nonexistent"})

	// Should not return error (writes structured error to stderr)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSetDefault(t *testing.T) {
	setupTestDir(t)

	credential.Store("proj1", credential.Credential{ConsoleKey: "k1"})
	credential.Store("proj2", credential.Credential{ConsoleKey: "k2"})
	config.StoreProject("proj1", config.Project{})
	config.StoreProject("proj2", config.Project{})

	root := &cobra.Command{Use: "test"}
	Register(root)
	root.SetArgs([]string{"project", "set-default", "proj2"})
	root.Execute()

	config.ClearCache()
	cfg := config.Read()
	if cfg.DefaultProject != "proj2" {
		t.Errorf("default = %q, want proj2", cfg.DefaultProject)
	}
}
