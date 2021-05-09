## Semantic version generator

Project created overnight, to prove that management of semantic versioning is NOT painful and do not require arguments and debates within the team. Simple, clean and only thing the project team should need to agree to are the keywords.

- [Semantic version generator](#semantic-version-generator)
  - [How does it work](#how-does-it-work)
  - [Usage](#usage)
    - [As a binary](#as-a-binary)
    - [As a github action](#as-a-github-action)
    - [Calculations example](#calculations-example)
    - [Example configuration](#example-configuration)
  - [Good to know](#good-to-know)

### How does it work

* Binary clones the github repository
* Iterates through the list of commits looking for the keywords specified in config file for additional bumps of versions
* Returns the semantic version which can be included in the release

### Usage

#### As a binary

```bash
bash$ ./semver-gen generate -r https://github.com/nextapps-de/winbox
SEMVER 9.0.10
bash$ ./semver-gen generate -l
SEMVER 5.1.1
```

** Local repository flag `-l` will always take precedence over remote repository URL **

```yaml
Usage:
  semver-gen generate [flags]
  semver-gen [command]

Available Commands:
  generate    Generates semantic version
  help        Help about any command

Flags:
  -c, --config string       Path to config file (default "config.yaml")
  -h, --help                help for semver-gen
  -l, --local               Use local repository
  -r, --repository string   Remote repository URL. (default "https://github.com/lukaszraczylo/simple-gql-client")
```

#### As a github action

```yaml
jobs:
  obain-semver:
    runs-on: ubuntu-latest
    name: Generate semver
    outputs:
      RELEASE_VERSION: ${{ steps.semver.outputs.SEMVER }}
    steps:
      - name: Run
        uses: lukaszraczylo/semver-generator
        id: semver
        with:
          config_file: config.yaml
          # then...
          repository_url: https://github.com/lukaszraczylo/simple-gql-client
          # or...
          repository_local: true
```

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

* version: is not respected at the moment, introduced for potential backwards compatibility in future
* force: sets the "starting" version, you don't need to specify this section as the default is always `0`
* wording: words the program should look for in the git commits to increment (patch|minor|major)

### Good to know

* Word matching uses fuzzy search AND is case INSENSITIVE
* I do not recommend using common words ( like "the" from the example configuration )