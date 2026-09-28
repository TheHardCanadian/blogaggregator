package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

//`json:"db_url"` is a struct tag. Without struct tag, Go would look for a JSON key matching the field name.
// Backticks (` `) are used in GO for attaching metadata to a struct field.

func Read() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(filepath.Join(home, ".gatorconfig.json"))
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg *Config) SetUser(username string) error {

	cfg.CurrentUserName = username

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "	")
	if err != nil {
		return err
	}

	path := filepath.Join(home, ".gatorconfig.json")
	return os.WriteFile(path, data, 0600)
}
