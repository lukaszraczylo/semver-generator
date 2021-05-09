#!/bin/sh -l
set -o pipefail

FLAGS=""

if [[ -z "$INPUT_CONFIG_FILE" ]]; then
  echo "Set the configuration file path."
  exit 1
else
  FLAGS="${FLAGS} -c $INPUT_CONFIG_FILE"
fi

if [[ -z "$INPUT_REPOSITORY_URL" ]] && [[ -z "$INPUT_REPOSITORY_LOCAL" ]];
then
  echo "You need to set either remote repository or repository local flags."
  exit 1
fi

if [[ ! -z "$INPUT_REPOSITORY_URL" ]]; then
  FLAGS="${FLAGS} -r $INPUT_REPOSITORY_URL"
fi

if [[ ! -z "$INPUT_REPOSITORY_LOCAL" ]]; then
  FLAGS="${FLAGS} -l"
fi

OUT_SEMVER_GEN=$(/go/src/app/semver-gen generate $FLAGS $*)
echo "::set-output name=semantic_version::$(echo $OUT_SEMVER_GEN | sed -e 's|SEMVER ||g')"
echo $OUT_SEMVER_GEN