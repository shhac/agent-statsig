package credential

import (
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
