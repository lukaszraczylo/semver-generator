package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

// Wording represents the keywords to look for in commit messages
type Wording struct {
	Patch   []string
	Minor   []string
	Major   []string
	Release []string
}

// Force represents forced versioning settings
type Force struct {
	Commit   string
	Patch    int
	Minor    int
	Major    int
	Existing bool
	Strict   bool
}

// Config represents the application configuration
type Config struct {
	Wording     Wording
	Force       Force
	Blacklist   []string
	TagPrefixes []string // Prefixes to strip from tags before parsing (e.g., "app-", "infra-", "v")
}

// ReadConfig reads the configuration from a file
func ReadConfig(file string) (*Config, error) {
	config := &Config{}

	viper.SetConfigFile(file)
	err := viper.ReadInConfig()
	if err != nil {
		err = fmt.Errorf("fatal error config file: %s", err)
		return config, err
	}

	if err := viper.UnmarshalKey("wording", &config.Wording); err != nil {
		return config, fmt.Errorf("error parsing wording config: %w", err)
	}
	if err := viper.UnmarshalKey("force", &config.Force); err != nil {
		return config, fmt.Errorf("error parsing force config: %w", err)
	}
	if err := viper.UnmarshalKey("blacklist", &config.Blacklist); err != nil {
		return config, fmt.Errorf("error parsing blacklist config: %w", err)
	}
	if err := viper.UnmarshalKey("tag_prefixes", &config.TagPrefixes); err != nil {
		return config, fmt.Errorf("error parsing tag_prefixes config: %w", err)
	}

	return config, nil
}

// ApplyForcedVersioning applies forced versioning settings to a semantic version
func ApplyForcedVersioning(force Force, semver *SemVer) {
	if force.Major > 0 {
		Debug("Forced versioning (MAJOR)", map[string]interface{}{"major": force.Major})
		semver.Major = force.Major
	}

	if force.Minor > 0 {
		Debug("Forced versioning (MINOR)", map[string]interface{}{"minor": force.Minor})
		semver.Minor = force.Minor
	}

	if force.Patch > 0 {
		Debug("Forced versioning (PATCH)", map[string]interface{}{"patch": force.Patch})
		semver.Patch = force.Patch
	}
}
