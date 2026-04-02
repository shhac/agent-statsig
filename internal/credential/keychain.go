package credential

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

const keychainService = "app.paulie.agent-statsig"

// keychainStore saves credentials to the macOS Keychain.
// Returns nil on success, non-nil if not on macOS or keychain operation fails.
func keychainStore(name, consoleKey, clientKey string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("keychain not available")
	}

	data, _ := json.Marshal(map[string]string{
		"console_key": consoleKey,
		"client_key":  clientKey,
	})

	// Delete existing entry if present (ignore errors).
	_ = exec.Command("security", "delete-generic-password", "-s", keychainService, "-a", name).Run()

	return exec.Command("security", "add-generic-password",
		"-s", keychainService, "-a", name, "-w", string(data),
		"-U",
	).Run()
}

// keychainGet retrieves credentials from the macOS Keychain.
func keychainGet(name string) (consoleKey, clientKey string, err error) {
	if runtime.GOOS != "darwin" {
		return "", "", fmt.Errorf("keychain not available")
	}

	out, err := exec.Command("security", "find-generic-password",
		"-s", keychainService, "-a", name, "-w",
	).Output()
	if err != nil {
		return "", "", err
	}

	var keys map[string]string
	if err := json.Unmarshal(out, &keys); err != nil {
		return "", "", err
	}
	return keys["console_key"], keys["client_key"], nil
}

// keychainDelete removes credentials from the macOS Keychain.
func keychainDelete(name string) {
	if runtime.GOOS != "darwin" {
		return
	}
	_ = exec.Command("security", "delete-generic-password", "-s", keychainService, "-a", name).Run()
}
