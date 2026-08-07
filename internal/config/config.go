package config

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/shhac/lib-agent-cli/creds"
	"github.com/shhac/lib-agent-cli/xdg"
)

type Config struct {
	DefaultProject string             `json:"default_project,omitempty"`
	Projects       map[string]Project `json:"projects"`
	Settings       Settings           `json:"settings"`
}

type Project struct {
	ConsoleKey string `json:"console_key,omitempty"`
	ClientKey  string `json:"client_key,omitempty"`
}

type Settings struct {
	Defaults *DefaultsSettings `json:"defaults,omitempty"`
}

type DefaultsSettings struct {
	Format string `json:"format,omitempty"`
}

var (
	cache       *Config
	cacheMu     sync.Mutex
	overrideDir string
)

// SetConfigDir overrides the config directory (for testing).
func SetConfigDir(dir string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	overrideDir = dir
	cache = nil
}

func ConfigDir() string {
	if overrideDir != "" {
		return overrideDir
	}
	return xdg.ConfigDir("agent-statsig")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// store is config.json's file: 0600 writes into a 0700 parent, atomic
// replacement, and Update for a locked read-modify-write. This used to be
// hand-rolled with os.ReadFile/os.WriteFile, which carried a lost-update
// race — two concurrent CLI invocations (e.g. `project add` racing
// `project set-default`) could each build their write from a snapshot taken
// before the other landed, silently erasing one of them.
func store() creds.Store {
	return creds.Store{Path: configPath()}
}

func Read() *Config {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if cache != nil {
		return cache
	}
	cache = loadConfig()
	return cache
}

// loadConfig reads config.json fresh from disk, bypassing the package cache.
// It is the single definition of "what a from-scratch read looks like",
// shared by Read (cached) and updateConfig (which must never hand a mutate
// callback the stale in-memory cache while holding the store's lock).
func loadConfig() *Config {
	var cfg Config
	if err := store().Load(&cfg); err != nil {
		return defaultConfig()
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]Project)
	}
	return &cfg
}

func Write(cfg *Config) error {
	if err := store().Save(cfg); err != nil {
		return err
	}
	cacheMu.Lock()
	cache = nil
	cacheMu.Unlock()
	return nil
}

func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = nil
}

func defaultConfig() *Config {
	return &Config{Projects: make(map[string]Project)}
}

// errSkipWrite lets a mutate callback decline to persist anything (e.g.
// SetDefault on an unknown alias) without updateConfig treating it as a real
// failure.
var errSkipWrite = errors.New("config: skip write")

// updateConfig applies mutate to a freshly loaded config under ONE exclusive
// lock spanning read, mutate, and write, so two concurrent invocations
// serialize instead of each building its write from a stale snapshot. The
// package-level cache is bypassed entirely while the lock is held — mutate
// always sees what store().Update just loaded from disk, never the cache —
// and is invalidated afterward so a later Read() cannot hand back the
// pre-write value.
func updateConfig(mutate func(cfg *Config) error) error {
	var cfg Config
	err := store().Update(&cfg, func() error {
		if cfg.Projects == nil {
			cfg.Projects = make(map[string]Project)
		}
		return mutate(&cfg)
	})

	cacheMu.Lock()
	cache = nil
	cacheMu.Unlock()

	if errors.Is(err, errSkipWrite) {
		return nil
	}
	return err
}

func StoreProject(alias string, proj Project) error {
	return updateConfig(func(cfg *Config) error {
		cfg.Projects[alias] = proj
		if cfg.DefaultProject == "" {
			cfg.DefaultProject = alias
		}
		return nil
	})
}

func RemoveProject(alias string) error {
	return updateConfig(func(cfg *Config) error {
		delete(cfg.Projects, alias)
		if cfg.DefaultProject == alias {
			cfg.DefaultProject = ""
			for name := range cfg.Projects {
				cfg.DefaultProject = name
				break
			}
		}
		return nil
	})
}

func SetDefault(alias string) error {
	return updateConfig(func(cfg *Config) error {
		if _, ok := cfg.Projects[alias]; !ok {
			return errSkipWrite
		}
		cfg.DefaultProject = alias
		return nil
	})
}
