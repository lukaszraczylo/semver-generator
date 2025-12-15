## Semantic version generator

[![Docker image build.](https://github.com/lukaszraczylo/semver-generator/actions/workflows/release.yaml/badge.svg)](https://github.com/lukaszraczylo/semver-generator/actions/workflows/release.yaml) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/lukaszraczylo/semver-generator) [![codecov](https://codecov.io/gh/lukaszraczylo/semver-generator/branch/main/graph/badge.svg?token=FY9BKETB59)](https://codecov.io/gh/lukaszraczylo/semver-generator)

Project created overnight, to prove that management of semantic versioning is NOT painful and do not require arguments and debates within the team. Simple, clean and only thing the project team should need to agree to are the keywords.

- [Semantic version generator](#semantic-version-generator)
  - [How does it work](#how-does-it-work)
  - [Additional features](#additional-features)
  - [Important changes](#important-changes)
  - [Usage](#usage)
    - [Authentication](#authentication)
    - [As a binary](#as-a-binary)
    - [As a github action](#as-a-github-action)
    - [As a docker container](#as-a-docker-container)
    - [Calculations example \[standard\]](#calculations-example-standard)
    - [Calculations example \[strict matching\]](#calculations-example-strict-matching)
    - [Release candidates](#release-candidates)
    - [Tag prefix stripping](#tag-prefix-stripping)
    - [Example configuration](#example-configuration)
  - [Good to know](#good-to-know)

### How does it work

* Binary clones the github repository
* Iterates through the list of commits looking for the keywords specified in config file for additional bumps of versions
* Returns the semantic version which can be included in the release

### Additional features

* With flag `-e` or config `force.existing: true` the existing tags in versioning will be respected, helping you to avoid the version conflicts.
* With config `force.commit: deadbeef` where `deadbeef` is the commit hash - calculations will start from the specified commit.
* Tag prefix stripping: The `v` prefix is automatically stripped from tags (e.g., `v1.2.3` → `1.2.3`). Additional prefixes can be configured via `tag_prefixes` for monorepo setups (e.g., `app-1.2.3`, `infra-1.2.3`).

### Important changes

* From version `1.4.2+` as pointed out in [issue #12](https://github.com/lukaszraczylo/semver-generator/issues/12) commits from merge will not be included in the calculations and commits themselves will bump the version on first match ( starting checks from `patch` upwards ).
* Added support for blacklisting terms to ignore specific commits, branch names, and merge messages from version calculations.

### Usage

#### Authentication

If you intend to use this project with remote repositories ( regardless of them being private or public ) you need to authenticate with your repository.
To do so you can utilise the following environment variables ( they are NOT github specific, for other providers you can use the password )

```bash
export GITHUB_USERNAME=lukaszraczylo
export GITHUB_TOKEN=yourPersonalApiToken
```

#### As a binary

##### Homebrew (macOS)

```bash
brew install --cask lukaszraczylo/taps/semver-generator
```

##### Manual Download

You can download latest versions of the binaries from the [release page](https://github.com/lukaszraczylo/semver-generator/releases/latest).

**Supported OS and architectures:**
Darwin ARM64/AMD64, Linux ARM64/AMD64, Windows AMD64

```bash
bash$ semver-generator generate -r https://github.com/nextapps-de/winbox
SEMVER 9.0.10
bash$ semver-generator generate -l
SEMVER 5.1.1
```

**Local repository flag `-l` will always take precedence over remote repository URL**

```yaml
Usage:
  semver-generator generate [flags]
  semver-generator [command]

Available Commands:
  generate    Generates semantic version
  help        Help about any command

Flags:
  -c, --config string       Path to config file (default "semver.yaml")
  -d, --debug               Enable debug mode
  -e, --existing            Respect existing tags
  -h, --help                help for semver-generator
  -l, --local               Use local repository
  -r, --repository string   Remote repository URL. (default "https://github.com/lukaszraczylo/simple-gql-client")
  -b, --branch string       Remote repository URL Branch. (default "main")
  -s, --strict              Strict matching
  -u, --update              Update binary with latest (no authentication required)
  -v, --version             Display version
```

##### Self-Update

The binary can update itself to the latest version:

```bash
semver-generator -u
```

This downloads the latest release for your platform directly from GitHub releases. No authentication is required.

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
        uses: lukaszraczylo/semver-generator@PLACE_LATEST_TAG_HERE
        # you can also use v1 tag which _should_ automatically upgrade to latest
        # uses: lukaszraczylo/semver-generator@v1
        with:
          config_file: semver.yaml
          # either...
          repository_local: true
          # or...
          repository_url: https://github.com/lukaszraczylo/simple-gql-client
          # when using remote repository, especially with private one:
          github_username: lukaszraczylo
          github_token: MySupeRSecr3tPa$$w0rd
          strict: true
          existing: false
      - name: Semver check
        run: |
          echo "Semantic version detected: ${{ steps.semver.outputs.semantic_version }}"
```

#### As a docker container

```bash
docker pull ghcr.io/lukaszraczylo/semver-generator:latest
```

#### Verifying Release Signatures

All release checksums and Docker images are signed with [cosign](https://github.com/sigstore/cosign) using keyless signing. To verify:

```bash
# Verify checksum signature
cosign verify-blob \
  --certificate-identity-regexp "https://github.com/lukaszraczylo/semver-generator/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  --bundle "<checksums-file>.sigstore.json" \
  <checksums-file>

# Verify Docker image
cosign verify \
  --certificate-identity-regexp "https://github.com/lukaszraczylo/semver-generator/.*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/lukaszraczylo/semver-generator:latest
```

**Docker supported architectures:**
Linux/arm64, Linux/amd64

#### Calculations example [standard]

```bash
- 0.0.1 - PATCH - starting commit
- 0.0.2 - PATCH - another commit
- 0.0.4 - PATCH - another commit with word 'Update' => DOUBLE increment PATCH
- 0.1.0 - MINOR - after commit with word 'Change' => increment MINOR, reset PATCH
- 0.1.1 - PATCH - additional commit
- 1.0.1 - MAJOR - commit with word 'BREAKING' = > INCREMENT MAJOR, reset MINOR
- 1.0.2 - PATCH - another commit
```

#### Calculations example [strict matching]

```bash
- 0.0.1 - PATCH - starting commit
- 0.0.1 - PATCH - another commit
- 0.0.1 - PATCH - another commit with word 'Update' => SINGLE increment PATCH
- 0.1.0 - MINOR - after commit with word 'Change' => increment MINOR, reset PATCH
- 0.1.0 - PATCH - additional commit
- 1.0.0 - MAJOR - commit with word 'BREAKING' = > INCREMENT MAJOR, reset MINOR
- 1.0.0 - PATCH - another commit
```

#### Release candidates

The `semver-gen` supports release candidates generation as well. Add following configuration ( and change the trigger keywords to anything what suits you )
to generate the appropriate release in format `1.3.37-rc.1` and counting up until next `minor` trigger will be detected.

```yaml
  release:
    - release-candidate
    - add-rc
```

#### Tag prefix stripping

When using the `-e` (existing tags) flag, the semver-generator needs to parse existing git tags to determine the current version. Tags often include prefixes that need to be stripped before version parsing.

**Automatic `v` prefix stripping:**
The `v` prefix is always stripped automatically from tags. For example:
- `v1.2.3` → parsed as `1.2.3`
- `v0.5.0` → parsed as `0.5.0`

**Custom prefixes for monorepos:**
In monorepo setups where different components have their own versioned tags, you can configure additional prefixes to strip:

```yaml
tag_prefixes:
  - "app-"
  - "infra-"
  - "api-"
  - "frontend-"
```

With this configuration:
- `app-1.2.3` → parsed as `1.2.3`
- `infra-0.5.0` → parsed as `0.5.0`
- `api-2.0.0-rc.1` → parsed as `2.0.0-rc.1` (release candidate)

This is particularly useful when:
- You have multiple services/components in a single repository
- Your CI/CD creates tags with component prefixes
- You want to track versions separately for different parts of your codebase

#### Example configuration

```yaml
version: 1
force:
  major: 1
  minor: 0
  patch: 1
  commit: 69fbe2df696f40281b9104ff073d26186cde1024
blacklist:
  - "Merge branch"
  - "Merge pull request"
  - "feature/"
  - "feature:"
tag_prefixes:
  - "app-"
  - "infra-"
  - "service-"
wording:
  patch:
    - update
    - initial
  minor:
    - change
    - improve
  major:
    - breaking
    - the # For testing purposes
  release:
    - release-candidate
    - add-rc
```

* `version`: is not respected at the moment, introduced for potential backwards compatibility in future
* `force`: sets the "starting" version, you don't need to specify this section as the default is always `0`
* `force.commit`: allows you to set commit hash from which the calculations should start
* `blacklist`: terms to ignore when processing commits. Any commit containing these terms will be skipped in version calculations. Useful for ignoring merge commits, feature branch names, and other unwanted triggers.
* `tag_prefixes`: prefixes to strip from existing tags before parsing version numbers. Useful for monorepos where tags are prefixed with component names (e.g., `app-1.2.3`, `infra-0.5.0`). The `v` prefix is always stripped automatically.
* `wording`: words the program should look for in the git commits to increment (patch|minor|major)

### Good to knows

* Word matching uses fuzzy search AND is case INSENSITIVE
* I do not recommend using common words ( like "the" from the example configuration )
* You can specify env variable `LOG_LEVEL=debug` to see what exactly happens during the calculations
