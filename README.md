## Semantic version generator

[![Docker image build.](https://github.com/lukaszraczylo/semver-generator/actions/workflows/release.yaml/badge.svg)](https://github.com/lukaszraczylo/semver-generator/actions/workflows/release.yaml) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/lukaszraczylo/semver-generator?style=plastic) [![Maintainability](https://api.codeclimate.com/v1/badges/0953eb7b7717af41911b/maintainability)](https://codeclimate.com/github/lukaszraczylo/semver-generator/maintainability) [![Test Coverage](https://api.codeclimate.com/v1/badges/0953eb7b7717af41911b/test_coverage)](https://codeclimate.com/github/lukaszraczylo/semver-generator/test_coverage)

Project created overnight, to prove that management of semantic versioning is NOT painful and do not require arguments and debates within the team. Simple, clean and only thing the project team should need to agree to are the keywords.

- [Semantic version generator](#semantic-version-generator)
  - [How does it work](#how-does-it-work)
  - [Usage](#usage)
    - [As a binary](#as-a-binary)
    - [As a github action](#as-a-github-action)
    - [As a docker container](#as-a-docker-container)
    - [Calculations example](#calculations-example)
    - [Example configuration](#example-configuration)
  - [Good to know](#good-to-know)

### How does it work

* Binary clones the github repository
* Iterates through the list of commits looking for the keywords specified in config file for additional bumps of versions
* Returns the semantic version which can be included in the release

### Usage

#### As a binary

You can download latest versions of the binaries from the [release page](https://github.com/lukaszraczylo/semver-generator/releases/latest).

**Supported OS and architectures:**
Darwin ARM64/AMD64, Linux ARM64/AMD64, Windows AMD64

```bash
bash$ ./semver-gen generate -r https://github.com/nextapps-de/winbox
SEMVER 9.0.10
bash$ ./semver-gen generate -l
SEMVER 5.1.1
```

**Local repository flag `-l` will always take precedence over remote repository URL**

```yaml
Usage:
  semver-gen generate [flags]
  semver-gen [command]

Available Commands:
  generate    Generates semantic version
  help        Help about any command

Flags:
  -c, --config string       Path to config file (default "config.yaml")
  -d, --debug               Enable debug mode
  -h, --help                help for semver-gen
  -l, --local               Use local repository
  -r, --repository string   Remote repository URL. (default "https://github.com/lukaszraczylo/simple-gql-client")
  -v, --version             Display version
```

#### As a github action

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
        uses: lukaszraczylo/semver-generator@1.0.29
        with:
          config_file: config.yaml
          # either...
          repository_local: true
          # or...
          repository_url: https://github.com/lukaszraczylo/simple-gql-client
      - name: Semver check
        run: |
          echo "Semantic version detected: ${{ steps.semver.outputs.semantic_version }}"
```

#### As a docker container

```bash
docker pull ghcr.io/lukaszraczylo/semver-generator:latest
```

**Docker supported architectures:**
Linux/arm64, Linux/amd64

#### Calculations example

* 0.0.1 - PATCH - starting commit
* 0.0.2 - PATCH - another commit
* 0.0.4 - PATCH - another commit with word 'Update' => DOUBLE increment PATCH
* 0.1.0 - MINOR - after commit with word 'Change' => increment MINOR, reset PATCH
* 0.1.1 - PATCH - additional commit
* 1.0.1 - MAJOR - commit with word 'BREAKING' = > INCREMENT MAJOR, reset MINOR
* 1.0.2 - PATCH - another commit

#### Example configuration

```yaml
version: 1
force:
  major: 1
  minor: 0
  patch: 1
  commit: 69fbe2df696f40281b9104ff073d26186cde1024
wording:
  patch:
    - update
    - initial
  minor:
    - add
    - change
    - improve
  major:
    - breaking
    - the # For testing purposes
```

* `version`: is not respected at the moment, introduced for potential backwards compatibility in future
* `force`: sets the "starting" version, you don't need to specify this section as the default is always `0`
* `force.commit`: allows you to set commit hash from which the calculations should start
* `wording`: words the program should look for in the git commits to increment (patch|minor|major)

### Good to know

* Word matching uses fuzzy search AND is case INSENSITIVE
* I do not recommend using common words ( like "the" from the example configuration )