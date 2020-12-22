#!/bin/bash

set -euo pipefail

ARTIFACT_CLI_VERSION="v0.2.8"
ARTIFACT_CLI_URL="https://github.com/semaphoreci/artifact/releases/download/$ARTIFACT_CLI_VERSION"

mkdir /tmp/Linux/toolbox
mkdir /tmp/Darwin/toolbox

#
# Get artifact releases
#
curl -s -L --retry 5 $ARTIFACT_CLI_URL/artifact_Linux_x86_64.tar.gz -o /tmp/Linux/Linux.tar.gz
curl -s -L --retry 5 $ARTIFACT_CLI_URL/artifact_Darwin_x86_64.tar.gz -o /tmp/Darwin/Darwin.tar.gz

#
# Unpack artifacts
#
cd /tmp/Linux && tar -zxf Linux.tar.gz && mv artifact toolbox/ && cd -
cd /tmp/Darwin && tar -zxf Darwin.tar.gz && mv artifact toolbox/ && cd -

#
# Cp toolbox files, to not mess up workspace
#
cp $SEMAPHORE_PROJECT_DIR/* /tmp/Linux/toolbox
cp $SEMAPHORE_PROJECT_DIR/* /tmp/Darwin/toolbox

#
# Create linux release
#
cd /tmp/Linux
tar -cf linux.tar toolbox

echo "toolbox Linux content: "
tar --list --verbose --file=linux.tar

#
# Create darwin release
#
cd /tmp/Darwin
tar -cf darwin.tar toolbox

echo "toolbox Darwin content: "
tar --list --verbose --file=darwin.tar

#
# Upload created release files to GitHub
#
curl \
  -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/semaphoreci/toolbox/releases \
  -d '{"tag_name":"'$SEMAPHORE_GIT_TAG_NAME'"}'

release_id=$(curl --silent https://api.github.com/repos/semaphoreci/toolbox/releases/tags/$SEMAPHORE_GIT_TAG_NAME | grep -m1 'id' | awk '{print $2}' | tr -d ',' )

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type /tmp/Linux/linux.tar)" \
    --data-binary @/tmp/Linux/linux.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=linux.tar"

echo "linux.tar uploaded"

curl \
    -X POST \
    -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Content-Type: $(file -b --mime-type /tmp/Darwin/darwin.tar)" \
    --data-binary @/tmp/Darwin/darwin.tar \
    "https://uploads.github.com/repos/semaphoreci/toolbox/releases/$release_id/assets?name=darwin.tar"

echo "darwin.tar uploaded"