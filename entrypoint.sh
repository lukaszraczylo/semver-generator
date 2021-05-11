#!/bin/sh -l
set -o pipefail

FLAGS=""

if [[ -z "$INPUT_CONFIG_FILE" ]]; then
  echo "Set the configuration file path."
else
  FLAGS="${FLAGS} -c $INPUT_CONFIG_FILE"
fi

if [[ -z "$INPUT_REPOSITORY_URL" ]] && [[ -z "$INPUT_REPOSITORY_LOCAL" ]];
then
  echo "You need to set either remote repository or repository local flags."
fi

if [[ ! -z "$INPUT_REPOSITORY_URL" ]]; then
  FLAGS="${FLAGS} -r $INPUT_REPOSITORY_URL"
fi

if [[ ! -z "$INPUT_REPOSITORY_LOCAL" ]]; then
  FLAGS="${FLAGS} -l"
fi

if [[ "${FLAGS}" == "" && "$*" == "" ]]; then
  exit 1
fi

cd /github/workspace
OUT_SEMVER_GEN=$(/go/src/app/semver-gen generate $FLAGS $*)
[ $? -eq 0 ] || exit 1
echo "::set-output name=semantic_version::$(echo $OUT_SEMVER_GEN | sed -e 's|SEMVER ||g')"
echo $OUT_SEMVER_GEN