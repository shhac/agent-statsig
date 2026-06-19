package credential

import (
	"encoding/json"
	"errors"

	"github.com/shhac/lib-agent-cli/creds"
)

// keychainService is owned by this CLI: the lib must not know the reverse-domain prefix.
const keychainService = "app.paulie.agent-statsig"

var keychain = creds.NewKeychain(keychainService)

// keychainStore saves credentials to the macOS Keychain.
// Returns nil on success, non-nil if not on macOS or keychain operation fails.
func keychainStore(name, consoleKey, clientKey string) error {
	if !keychain.Available() {
		return creds.ErrKeychainUnavailable
	}

	data, _ := json.Marshal(map[string]string{
		"console_key": consoleKey,
		"client_key":  clientKey,
	})

	return keychain.Set(name, string(data))
}

// keychainGet retrieves credentials from the macOS Keychain.
func keychainGet(name string) (consoleKey, clientKey string, err error) {
	if !keychain.Available() {
		return "", "", creds.ErrKeychainUnavailable
	}

	value, ok := keychain.Get(name)
	if !ok {
		return "", "", errors.New("keychain entry not found")
	}

	var keys map[string]string
	if err := json.Unmarshal([]byte(value), &keys); err != nil {
		return "", "", err
	}
	return keys["console_key"], keys["client_key"], nil
}

// keychainDelete removes credentials from the macOS Keychain.
func keychainDelete(name string) {
	if !keychain.Available() {
		return
	}
	_ = keychain.Delete(name)
}
