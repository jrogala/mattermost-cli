// Package config handles CLI configuration from env vars, OS keyring, and config files.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
)

const keyringService = "mattermost-cli"

// Config holds the Mattermost connection settings.
type Config struct {
	URL           string `mapstructure:"url"`
	Token         string `mapstructure:"-"` // not stored in config file
	TLSSkipVerify bool   `mapstructure:"tls_skip_verify"`
}

// Load reads config with the following priority:
//  1. Env vars: MATTERMOST_URL, MATTERMOST_TOKEN, MATTERMOST_TLS_SKIP_VERIFY
//  2. OS keyring for token
//  3. Config file (~/.config/mattermost-cli/config.yaml)
//  4. Config file token field (legacy fallback)
func Load() (*Config, error) {
	viper.SetEnvPrefix("MATTERMOST")
	viper.AutomaticEnv()

	configDir, err := os.UserConfigDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(configDir, "mattermost-cli"))
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	_ = viper.ReadInConfig()

	cfg := &Config{
		URL:           viper.GetString("url"),
		TLSSkipVerify: viper.GetBool("tls_skip_verify"),
	}

	// Token resolution: env var > keyring > config file (legacy)
	cfg.Token = viper.GetString("token")
	if envToken := os.Getenv("MATTERMOST_TOKEN"); envToken != "" {
		cfg.Token = envToken
	} else if t, err := keyring.Get(keyringService, "token"); err == nil && t != "" {
		cfg.Token = t
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("mattermost URL not set. Run 'mattermost-cli login' first")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("mattermost token not set. Run 'mattermost-cli login' first")
	}

	return cfg, nil
}

// ConfigDir returns the config directory path.
func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config dir: %w", err)
	}
	return filepath.Join(configDir, "mattermost-cli"), nil
}

// Save writes URL and settings to config file, token to OS keyring.
// Falls back to config file for token if keyring is unavailable.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("cannot create config dir: %w", err)
	}

	// Try to store token in OS keyring
	keyringOK := false
	if err := keyring.Set(keyringService, "token", cfg.Token); err == nil {
		keyringOK = true
	} else {
		fmt.Fprintf(os.Stderr, "Warning: OS keyring unavailable (%v), storing token in config file\n", err)
	}

	// Write config file: URL + settings, token only if keyring failed
	content := fmt.Sprintf("url: %s\n", cfg.URL)
	if cfg.TLSSkipVerify {
		content += "tls_skip_verify: true\n"
	}
	if !keyringOK {
		content += fmt.Sprintf("token: %s\n", cfg.Token)
	}

	return os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0600)
}
