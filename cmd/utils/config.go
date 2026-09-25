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
	Matching    string   // Effective matching mode: MatchingFuzzy (default) or MatchingRegex
	Matcher     KeywordMatcher
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

	// Read "matching" first, right after ReadInConfig, so every later parse
	// error below can tell whether regex mode is in play. viper.Get (not
	// GetString) lets us reject a non-scalar value, such as a YAML list or
	// map, as a config error instead of silently treating it as "fuzzy".
	switch rawMatching := viper.Get("matching"); v := rawMatching.(type) {
	case nil:
		config.Matching = MatchingFuzzy
	case string:
		switch v {
		case "":
			config.Matching = MatchingFuzzy
		case MatchingFuzzy, MatchingRegex:
			config.Matching = v
		default:
			return config, fmt.Errorf("%w: invalid matching value %q, allowed values are %q and %q", ErrInvalidMatchingConfig, v, MatchingFuzzy, MatchingRegex)
		}
	default:
		return config, fmt.Errorf("%w: matching config must be a string, got %v (%T)", ErrInvalidMatchingConfig, v, v)
	}

	// In regex mode there is no fuzzy fallback: any config parse error below
	// (wording/force/blacklist/tag_prefixes) is a real misconfiguration, so
	// wrap it with ErrInvalidMatchingConfig, which main.go treats as fatal.
	// In fuzzy mode (default), behaviour is unchanged: the error is returned
	// as-is and the caller logs it and continues with defaults.
	wrapIfRegex := func(err error) error {
		if err == nil || config.Matching != MatchingRegex {
			return err
		}
		return fmt.Errorf("%w: %v", ErrInvalidMatchingConfig, err)
	}

	if err := viper.UnmarshalKey("wording", &config.Wording); err != nil {
		return config, wrapIfRegex(fmt.Errorf("error parsing wording config: %w", err))
	}
	if err := viper.UnmarshalKey("force", &config.Force); err != nil {
		return config, wrapIfRegex(fmt.Errorf("error parsing force config: %w", err))
	}
	if err := viper.UnmarshalKey("blacklist", &config.Blacklist); err != nil {
		return config, wrapIfRegex(fmt.Errorf("error parsing blacklist config: %w", err))
	}
	if err := viper.UnmarshalKey("tag_prefixes", &config.TagPrefixes); err != nil {
		return config, wrapIfRegex(fmt.Errorf("error parsing tag_prefixes config: %w", err))
	}

	matcher, err := NewKeywordMatcher(config.Matching, config.Wording)
	if err != nil {
		return config, err
	}
	config.Matcher = matcher

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
