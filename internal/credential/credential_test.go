package credential

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shhac/agent-statsig/internal/config"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	config.SetConfigDir(dir)
	t.Cleanup(func() { config.SetConfigDir("") })
}

func TestStoreAndGet(t *testing.T) {
	setupTestDir(t)

	cred := Credential{
		ConsoleKey: "console-xxx",
		ClientKey:  "client-yyy",
	}
	storage, err := Store("myproject", cred)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if storage != "keychain" && storage != "file" {
		t.Errorf("unexpected storage type: %s", storage)
	}

	got, err := Get("myproject")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ConsoleKey != "console-xxx" {
		t.Errorf("ConsoleKey = %q", got.ConsoleKey)
	}
	if got.ClientKey != "client-yyy" {
		t.Errorf("ClientKey = %q", got.ClientKey)
	}
}

// TestStore_Headless_FileFallback exercises the real credential-WRITE path
// non-interactively. Setting the per-CLI keychain opt-out (derived by
// lib-agent-cli from the "app.paulie.agent-statsig" service) makes the keychain
// report Available()==false, so Store deterministically reports storage "file"
// and writes the real keys to the 0600 credentials file on every platform —
// including darwin, which would otherwise reach the `security` CLI and its GUI
// prompt. This is the first non-interactive coverage of the write path.
func TestStore_Headless_FileFallback(t *testing.T) {
	t.Setenv("AGENT_STATSIG_NO_KEYCHAIN", "1")
	dir := t.TempDir()
	config.SetConfigDir(dir)
	t.Cleanup(func() { config.SetConfigDir("") })

	cred := Credential{ConsoleKey: "console-headless", ClientKey: "client-headless"}
	storage, err := Store("headless", cred)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if storage != "file" {
		t.Fatalf("storage=%q, want \"file\" (keychain opt-out should force the file path)", storage)
	}

	// The keys must land in the 0600 file (no keychain sentinel, since nothing
	// was pushed to a keychain).
	path := filepath.Join(dir, "credentials.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("credentials file not written: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("credentials mode=%o, want 0600", mode)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !strings.Contains(string(raw), "console-headless") {
		t.Fatalf("file fallback expected the raw key on disk; got:\n%s", raw)
	}
	if strings.Contains(string(raw), keychainSentinel) {
		t.Fatalf("file unexpectedly contains keychain sentinel (keychain should be bypassed):\n%s", raw)
	}

	// Round-trip via the read path.
	got, err := Get("headless")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ConsoleKey != "console-headless" || got.ClientKey != "client-headless" {
		t.Fatalf("round-trip = %+v; want console-headless/client-headless", got)
	}

	if err := Remove("headless"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := Get("headless"); err == nil {
		t.Fatal("expected not found after Remove")
	}
}

func TestGetNotFound(t *testing.T) {
	setupTestDir(t)

	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing credential")
	}
	nf, ok := err.(*NotFoundError)
	if !ok {
		t.Fatalf("expected *NotFoundError, got %T", err)
	}
	if nf.Name != "nonexistent" {
		t.Errorf("NotFoundError.Name = %q", nf.Name)
	}
}

func TestRemove(t *testing.T) {
	setupTestDir(t)

	Store("toremove", Credential{ConsoleKey: "key"})

	if err := Remove("toremove"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	_, err := Get("toremove")
	if err == nil {
		t.Error("expected not found after remove")
	}
}

func TestRemoveNotFound(t *testing.T) {
	setupTestDir(t)

	err := Remove("nope")
	if err == nil {
		t.Fatal("expected error for removing nonexistent")
	}
	if _, ok := err.(*NotFoundError); !ok {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
}

func TestList(t *testing.T) {
	setupTestDir(t)

	Store("alpha", Credential{ConsoleKey: "k1"})
	Store("beta", Credential{ConsoleKey: "k2"})

	names, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("got %d names, want 2", len(names))
	}

	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	if !found["alpha"] || !found["beta"] {
		t.Errorf("missing expected names: %v", names)
	}
}

func TestNotFoundErrorMessage(t *testing.T) {
	err := &NotFoundError{Name: "test"}
	want := `project credential "test" not found`
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestStoreOverwrite(t *testing.T) {
	setupTestDir(t)

	Store("proj", Credential{ConsoleKey: "old"})
	Store("proj", Credential{ConsoleKey: "new", ClientKey: "client"})

	got, _ := Get("proj")
	if got.ConsoleKey != "new" {
		t.Errorf("ConsoleKey should be overwritten, got %q", got.ConsoleKey)
	}
	if got.ClientKey != "client" {
		t.Errorf("ClientKey = %q", got.ClientKey)
	}
}
