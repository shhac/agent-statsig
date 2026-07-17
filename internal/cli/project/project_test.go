package project

import (
	"strings"
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

	// Single-sink: the command bubbles a fixable error for the caller to render.
	if err := root.Execute(); err == nil {
		t.Fatal("expected error when --console-key is missing")
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

// TestAddConsoleKeyFromStdin covers the non-interactive secret path: the
// console key is piped on stdin (kept off argv/history) rather than passed
// as a flag. ReadSecret trims the piped value.
func TestAddConsoleKeyFromStdin(t *testing.T) {
	setupTestDir(t)

	root := &cobra.Command{Use: "test"}
	Register(root)
	root.SetArgs([]string{"project", "add", "stdinproj", "--client-key", "client-pub"})
	root.SetIn(strings.NewReader("  console-piped\n"))
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	cred, err := credential.Get("stdinproj")
	if err != nil {
		t.Fatal(err)
	}
	if cred.ConsoleKey != "console-piped" {
		t.Errorf("ConsoleKey = %q, want trimmed 'console-piped'", cred.ConsoleKey)
	}
	if cred.ClientKey != "client-pub" {
		t.Errorf("ClientKey = %q, want 'client-pub'", cred.ClientKey)
	}
}

// TestAddConsoleKeyFlagWinsOverStdin verifies precedence: an explicit
// --console-key short-circuits ReadSecret, so a differing piped value is
// never consulted.
func TestAddConsoleKeyFlagWinsOverStdin(t *testing.T) {
	setupTestDir(t)

	root := &cobra.Command{Use: "test"}
	Register(root)
	root.SetArgs([]string{"project", "add", "flagwins", "--console-key", "from-flag"})
	root.SetIn(strings.NewReader("from-stdin"))
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	cred, err := credential.Get("flagwins")
	if err != nil {
		t.Fatal(err)
	}
	if cred.ConsoleKey != "from-flag" {
		t.Errorf("ConsoleKey = %q, want 'from-flag' (flag must win over stdin)", cred.ConsoleKey)
	}
}

// TestAddEmptyStdinStillRequiresConsoleKey confirms the required-ness error
// still fires when neither the flag nor a piped stdin supplies the key.
func TestAddEmptyStdinStillRequiresConsoleKey(t *testing.T) {
	setupTestDir(t)

	root := &cobra.Command{Use: "test"}
	Register(root)
	root.SetArgs([]string{"project", "add", "emptyproj"})
	root.SetIn(strings.NewReader(""))

	if err := root.Execute(); err == nil {
		t.Fatal("expected error when neither --console-key nor stdin supplies the key")
	}

	if _, err := credential.Get("emptyproj"); err == nil {
		t.Error("credential should not exist without a console key")
	}
}

func TestRemoveNonexistent(t *testing.T) {
	setupTestDir(t)

	root := &cobra.Command{Use: "test"}
	Register(root)
	root.SetArgs([]string{"project", "remove", "nonexistent"})

	// Single-sink: removing a missing project bubbles a fixable error.
	if err := root.Execute(); err == nil {
		t.Fatal("expected error when removing a nonexistent project")
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
