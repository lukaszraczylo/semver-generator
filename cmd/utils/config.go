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
	Wording   Wording
	Force     Force
	Blacklist []string
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
	
	viper.UnmarshalKey("wording", &config.Wording)
	viper.UnmarshalKey("force", &config.Force)
	viper.UnmarshalKey("blacklist", &config.Blacklist)
	
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