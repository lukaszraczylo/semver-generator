# Semantic Version Generator

[![Docker image build.](https://github.com/lukaszraczylo/semver-generator/actions/workflows/release.yaml/badge.svg)](https://github.com/lukaszraczylo/semver-generator/actions/workflows/release.yaml) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/lukaszraczylo/semver-generator) [![codecov](https://codecov.io/gh/lukaszraczylo/semver-generator/branch/main/graph/badge.svg?token=FY9BKETB59)](https://codecov.io/gh/lukaszraczylo/semver-generator)

A lightweight, configurable tool that simplifies semantic versioning by automatically calculating version numbers based on git commit messages. No more manual version management or team debates about versioning - just agree on the keywords and let the tool handle the rest.

## Table of Contents

- [How It Works](#how-it-works)
- [Key Features](#key-features)
- [Important Changes](#important-changes)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Binary](#binary)
  - [Docker](#docker)
  - [GitHub Action](#github-action)
- [Usage](#usage)
  - [Authentication](#authentication)
  - [Command Line Options](#command-line-options)
  - [Configuration File](#configuration-file)
  - [Versioning Behavior](#versioning-behavior)
  - [Release Candidates](#release-candidates)
- [Examples](#examples)
  - [Standard Mode](#standard-mode)
  - [Strict Matching Mode](#strict-matching-mode)
- [Advanced Features](#advanced-features)
  - [Force Settings](#force-settings)
  - [Blacklist Terms](#blacklist-terms)
- [Tips & Best Practices](#tips--best-practices)

## How It Works

The semantic version generator follows a simple process:

1. Clones the specified GitHub repository (or uses a local repository)
2. Iterates through the commit history, analyzing each commit message
3. Looks for predefined keywords (configurable) that trigger version increments
4. Calculates the appropriate semantic version based on the matches
5. Returns the resulting version that can be used for releases

## Key Features

- **Effortless Version Calculation**: Automatically determine the appropriate version based on commit messages
- **Configurable Keywords**: Define your own keywords for patch, minor, and major version increments
- **Support for Existing Tags**: Option to respect existing version tags to avoid conflicts
- **Release Candidate Support**: Generate release candidate versions with incrementing counter (e.g., `1.2.3-rc.1`)
- **Flexible Repository Source**: Work with either local or remote Git repositories
- **Blacklist Support**: Ignore specific commits or branch merges from version calculations
- **Force Options**: Start calculations from a specific commit or set a minimum version

## Important Changes

- **Since v1.4.2+**: Commits from merge requests are no longer included in calculations
- **Commit Matching Behavior**: Commits will bump the version on the first match (checking from `patch` upwards)
- **Blacklist Support**: Added ability to ignore specific terms in commit messages, branch names, and merge requests

## Installation

### Prerequisites

When using with remote repositories, authentication is required. Set the following environment variables:

```bash
export GITHUB_USERNAME=yourusername
export GITHUB_TOKEN=yourpersonalapitoken
```

### Binary

Download the latest binary from the [release page](https://github.com/lukaszraczylo/semver-generator/releases/latest).

**Supported platforms**:
- Darwin (macOS): ARM64/AMD64
- Linux: ARM64/AMD64
- Windows: AMD64

### Docker

```bash
docker pull ghcr.io/lukaszraczylo/semver-generator:latest
```

**Supported architectures**:
- Linux/arm64
- Linux/amd64

### GitHub Action

Add to your GitHub workflow:

```yaml
jobs:
  prepare:
    name: Preparing build context
    runs-on: ubuntu-latest
    outputs:
      RELEASE_VERSION: ${{ steps.semver.outputs.semantic_version }}
    steps:
      - name: Checkout repo
        uses: actions/checkout@v2
        with:
          fetch-depth: '0'
      - name: Semver run
        id: semver
        uses: lukaszraczylo/semver-generator@v1
        with:
          config_file: semver.yaml
          # Either use local repository
          repository_local: true
          # Or specify remote repository
          # repository_url: https://github.com/lukaszraczylo/simple-gql-client
          # github_username: ${{ secrets.GH_USERNAME }}
          # github_token: ${{ secrets.GH_TOKEN }}
          strict: false
          existing: true
      - name: Use semantic version
        run: |
          echo "Semantic version detected: ${{ steps.semver.outputs.semantic_version }}"
```

## Usage

### Authentication

For remote repositories (public or private), authenticate using:

```bash
export GITHUB_USERNAME=yourusername
export GITHUB_TOKEN=yourpersonalapitoken
```

### Command Line Options

```
Usage:
  semver-gen generate [flags]
  semver-gen [command]

Available Commands:
  generate    Generates semantic version
  help        Help about any command

Flags:
  -c, --config string       Path to config file (default "semver.yaml")
  -d, --debug               Enable debug mode
  -e, --existing            Respect existing tags
  -h, --help                help for semver-gen
  -l, --local               Use local repository
  -r, --repository string   Remote repository URL. (default "https://github.com/lukaszraczylo/simple-gql-client")
  -b, --branch string       Remote repository URL Branch. (default "main")
  -s, --strict              Strict matching
  -u, --update              Update binary with latest
  -v, --version             Display version
```

**Note**: The `-l/--local` flag takes precedence over the repository URL.

### Configuration File

Create a `semver.yaml` file (or specify a different path with `-c`):

```yaml
version: 1
force:
  major: 1
  minor: 0
  patch: 1
  commit: 69fbe2df696f40281b9104ff073d26186cde1024
  existing: true
  strict: false
blacklist:
  - "Merge branch"
  - "Merge pull request"
  - "feature/"
  - "feature:"
wording:
  patch:
    - update
    - initial
    - fix
  minor:
    - change
    - improve
    - add
  major:
    - breaking
    - redesign
  release:
    - release-candidate
    - add-rc
```

Configuration options:
- `version`: Reserved for future backward compatibility
- `force`: Set starting version or other constraints
- `blacklist`: Terms to ignore when processing commits
- `wording`: Keywords that trigger version increments

### Versioning Behavior

The version calculation follows semantic versioning rules:
- **MAJOR**: Incremented for incompatible API changes
- **MINOR**: Incremented for backward-compatible new features
- **PATCH**: Incremented for backward-compatible bug fixes

When keywords are matched:
- **MAJOR** match: Increments major version, resets minor and patch versions
- **MINOR** match: Increments minor version, resets patch version
- **PATCH** match: Increments patch version
- By default (non-strict mode), every commit increments patch version

### Release Candidates

Add the following to your configuration to generate release candidates:

```yaml
wording:
  release:
    - release-candidate
    - add-rc
```

When a release candidate keyword is found, the version will be formatted as `1.2.3-rc.1`, with the counter incrementing for each new release candidate.

## Examples

### Standard Mode

```
- 0.0.1 - PATCH - starting commit
- 0.0.2 - PATCH - another commit
- 0.0.4 - PATCH - another commit with word 'Update' => DOUBLE increment PATCH
- 0.1.0 - MINOR - after commit with word 'Change' => increment MINOR, reset PATCH
- 0.1.1 - PATCH - additional commit
- 1.0.1 - MAJOR - commit with word 'BREAKING' => INCREMENT MAJOR, reset MINOR
- 1.0.2 - PATCH - another commit
```

### Strict Matching Mode

In strict mode (using `-s` flag or `force.strict: true`), versions only increment when a keyword is matched:

```
- 0.0.1 - PATCH - starting commit
- 0.0.1 - PATCH - another commit (no change - no keyword match)
- 0.0.2 - PATCH - another commit with word 'Update' => increment PATCH
- 0.1.0 - MINOR - after commit with word 'Change' => increment MINOR, reset PATCH
- 0.1.0 - PATCH - additional commit (no change - no keyword match)
- 1.0.0 - MAJOR - commit with word 'BREAKING' => INCREMENT MAJOR, reset MINOR
- 1.0.0 - PATCH - another commit (no change - no keyword match)
```

## Advanced Features

### Force Settings

Control versioning with force settings in configuration:

```yaml
force:
  major: 1  # Set minimum major version
  minor: 0  # Set minimum minor version
  patch: 1  # Set minimum patch version
  commit: 69fbe2df696f40281b9104ff073d26186cde1024  # Start from specific commit
  existing: true  # Respect existing tags (same as -e flag)
  strict: false  # Use strict matching (same as -s flag)
```

### Blacklist Terms

Ignore specific commits from version calculations:

```yaml
blacklist:
  - "Merge branch"  # Ignore merge commits
  - "Merge pull request"  # Ignore PR merges
  - "feature/"  # Ignore feature branch names
  - "chore:"  # Ignore chore commits
```

## Tips & Best Practices

- Word matching uses fuzzy search and is case INSENSITIVE
- Avoid common words as version triggers (e.g., "the", "and")
- Use `LOG_LEVEL=debug` environment variable to see detailed calculation steps
- When using as a GitHub Action, ensure `fetch-depth: '0'` to get the complete commit history
- For complex projects, consider using a more specific configuration to distinguish between types of changes
