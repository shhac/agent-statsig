package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
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
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "agent-statsig")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "agent-statsig")
}

func configPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func Read() *Config {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if cache != nil {
		return cache
	}
	data, err := os.ReadFile(configPath())
	if err != nil {
		return defaultConfig()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultConfig()
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]Project)
	}
	cache = &cfg
	return cache
}

func Write(cfg *Config) error {
	cacheMu.Lock()
	cache = nil
	cacheMu.Unlock()

	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), append(data, '\n'), 0o644)
}

func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = nil
}

func defaultConfig() *Config {
	cfg := &Config{
		Projects: make(map[string]Project),
	}
	cache = cfg
	return cfg
}

func StoreProject(alias string, proj Project) error {
	cfg := Read()
	cfg.Projects[alias] = proj
	if cfg.DefaultProject == "" {
		cfg.DefaultProject = alias
	}
	return Write(cfg)
}

func RemoveProject(alias string) error {
	cfg := Read()
	delete(cfg.Projects, alias)
	if cfg.DefaultProject == alias {
		cfg.DefaultProject = ""
		for name := range cfg.Projects {
			cfg.DefaultProject = name
			break
		}
	}
	return Write(cfg)
}

func SetDefault(alias string) error {
	cfg := Read()
	if _, ok := cfg.Projects[alias]; !ok {
		return nil
	}
	cfg.DefaultProject = alias
	return Write(cfg)
}
