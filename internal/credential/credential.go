package credential

import (
	"fmt"
	"path/filepath"

	"github.com/shhac/agent-statsig/internal/config"
	"github.com/shhac/lib-agent-cli/creds"
)

const keychainSentinel = "__KEYCHAIN__"

type Credential struct {
	ConsoleKey      string `json:"console_key"`
	ClientKey       string `json:"client_key,omitempty"`
	KeychainManaged bool   `json:"keychain_managed,omitempty"`
}

type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("project credential %q not found", e.Name)
}

func credentialsPath() string {
	return filepath.Join(config.ConfigDir(), "credentials.json")
}

// store is the credential index's file: 0600 writes into a 0700 parent, atomic
// replacement, and Update for a locked read-modify-write. This used to be
// hand-rolled with os.ReadFile/os.WriteFile, which carried a lost-update race —
// two concurrent writers could each build their write from a stale snapshot,
// and the loser's entry vanished while its secret stayed in the keychain,
// unreferenced and un-removable (auth list can't show it, auth remove can't
// look it up).
func store() creds.Store {
	return creds.Store{Path: credentialsPath()}
}

func readIndex() (map[string]Credential, error) {
	index := map[string]Credential{}
	if err := store().Load(&index); err != nil {
		return nil, err
	}
	if index == nil {
		index = map[string]Credential{}
	}
	return index, nil
}

// updateIndex applies mutate to the index under an exclusive lock, so two
// concurrent `project add`/`project remove` invocations serialize instead of
// clobbering each other.
func updateIndex(mutate func(index map[string]Credential) error) error {
	index := map[string]Credential{}
	return store().Update(&index, func() error {
		if index == nil {
			index = map[string]Credential{}
		}
		return mutate(index)
	})
}

func Store(name string, cred Credential) (string, error) {
	storage := "file"
	entry := Credential{
		ConsoleKey: cred.ConsoleKey,
		ClientKey:  cred.ClientKey,
	}

	if err := keychainStore(name, cred.ConsoleKey, cred.ClientKey); err == nil {
		entry.ConsoleKey = keychainSentinel
		entry.ClientKey = keychainSentinel
		entry.KeychainManaged = true
		storage = "keychain"
	}

	// The index write is the step that must not race: the keychain already
	// holds the secret by now, so an entry lost to a concurrent writer leaves
	// that secret referenced by nothing.
	if err := updateIndex(func(index map[string]Credential) error {
		index[name] = entry
		return nil
	}); err != nil {
		return "", err
	}
	return storage, nil
}

func Get(name string) (*Credential, error) {
	index, err := readIndex()
	if err != nil {
		return nil, err
	}
	entry, ok := index[name]
	if !ok {
		return nil, &NotFoundError{Name: name}
	}

	cred := entry
	if entry.KeychainManaged {
		if consoleKey, clientKey, err := keychainGet(name); err == nil {
			cred.ConsoleKey = consoleKey
			cred.ClientKey = clientKey
		}
	}

	return &cred, nil
}

func Remove(name string) error {
	return updateIndex(func(index map[string]Credential) error {
		entry, ok := index[name]
		if !ok {
			return &NotFoundError{Name: name}
		}

		if entry.KeychainManaged {
			keychainDelete(name)
		}

		delete(index, name)
		return nil
	})
}

func List() ([]string, error) {
	index, err := readIndex()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(index))
	for name := range index {
		names = append(names, name)
	}
	return names, nil
}
