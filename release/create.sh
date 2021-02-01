#!/bin/bash

set -euo pipefail

ARTIFACT_CLI_VERSION="v0.2.8"
WHEN_CLI_VERSION="v1.0.5"
SPC_CLI_VERSION="v1.1.0"

ARTIFACT_CLI_URL="https://github.com/semaphoreci/artifact/releases/download/$ARTIFACT_CLI_VERSION"
SPC_CLI_URL="https://github.com/semaphoreci/spc/releases/download/$SPC_CLI_VERSION"
WHEN_CLI_URL="https://github.com/renderedtext/when/releases/download/$WHEN_CLI_VERSION"

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

curl -s -L --fail --retry 5 $ARTIFACT_CLI_URL/artifact_Linux_x86_64.tar.gz -o /tmp/Linux/artifact_Linux.tar.gz
curl -s -L --fail --retry 5 $ARTIFACT_CLI_URL/artifact_Darwin_x86_64.tar.gz -o /tmp/Darwin/artifact_Darwin.tar.gz

cd /tmp/Linux && tar -zxf artifact_Linux.tar.gz && mv artifact toolbox/ && cd -
cd /tmp/Darwin && tar -zxf artifact_Darwin.tar.gz && mv artifact toolbox/ && cd -

#
# Download and add When CLI to the release
#
echo "Download When CLI"

curl -s -L --retry 5 $WHEN_CLI_URL/when -o /tmp/Linux/toolbox/when

chmod +x /tmp/Linux/toolbox/when

#
# Download and add SPC CLI to the release
#
echo "Download SPC CLI"

curl -s -L --fail --retry 5 $SPC_CLI_URL/spc_Linux_x86_64.tar.gz -o /tmp/Linux/spc_Linux.tar.gz

cd /tmp/Linux && tar -zxf spc_Linux.tar.gz && mv spc toolbox/ && cd -

#
# Create linux release
#
echo "Creating linux.tar"

cd /tmp/Linux
tar -cf linux.tar toolbox

echo "toolbox Linux content: "
tar --list --verbose --file=linux.tar

#
# Create mac release
#
echo "Creating darwin.tar"

cd /tmp/Darwin
tar -cf darwin.tar toolbox

echo "toolbox Darwin content: "
tar --list --verbose --file=darwin.tar
