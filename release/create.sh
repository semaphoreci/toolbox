#!/bin/bash

set -euo pipefail

ARTIFACT_CLI_VERSION="v0.4.0"
ARTIFACT_CLI_URL="https://github.com/semaphoreci/artifact/releases/download/$ARTIFACT_CLI_VERSION"

#
# Create release dirs
#
echo "Create initial content of the release directories"

rm -rf /tmp/Linux
rm -rf /tmp/Darwin

mkdir -p /tmp/Linux/toolbox
mkdir -p /tmp/Darwin/toolbox

cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Linux/toolbox
cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Darwin/toolbox

rm -rf /tmp/Linux/toolbox/.git
rm -rf /tmp/Darwin/toolbox/.git

rm -rf /tmp/Linux/toolbox/.gitignore
rm -rf /tmp/Darwin/toolbox/.gitignore

rm -rf /tmp/Linux/toolbox/Makefile
rm -rf /tmp/Darwin/toolbox/Makefile

rm -rf /tmp/Linux/toolbox/release
rm -rf /tmp/Darwin/toolbox/release

rm -rf /tmp/Linux/toolbox/tests
rm -rf /tmp/Darwin/toolbox/tests

#
# Download and add artifact CLI to the release
#
echo "Download Artifact CLI"

curl -s -L --retry 5 $ARTIFACT_CLI_URL/artifact_Linux_x86_64.tar.gz -o /tmp/Linux/artifact_Linux.tar.gz
curl -s -L --retry 5 $ARTIFACT_CLI_URL/artifact_Darwin_x86_64.tar.gz -o /tmp/Darwin/artifact_Darwin.tar.gz

cd /tmp/Linux && tar -zxf artifact_Linux.tar.gz && mv artifact toolbox/ && cd -
cd /tmp/Darwin && tar -zxf artifact_Darwin.tar.gz && mv artifact toolbox/ && cd -

#
# Create linux release
#
echo "Creating linux.tar"

cd /tmp/Linux
tar -cf linux.tar toolbox

echo "toolbox Linux content: "
tar --list --verbose --file=linux.tar

#
# Create darwin release
#
echo "Creating darwin.tar"

cd /tmp/Darwin
tar -cf darwin.tar toolbox

echo "toolbox Darwin content: "
tar --list --verbose --file=darwin.tar

#
# Upload created release files to GitHub
#

echo "Uploading to GitHub"

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

echo "Everything DONE"
echo "🍻"
