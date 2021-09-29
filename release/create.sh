#!/bin/bash

set -euo pipefail

ARTIFACT_CLI_VERSION="v0.4.6"
WHEN_CLI_VERSION="v1.0.5"
SPC_CLI_VERSION="v1.9.1"
TEST_RESULTS_CLI_VERSION="v0.4.10"

ARTIFACT_CLI_URL="https://github.com/semaphoreci/artifact/releases/download/$ARTIFACT_CLI_VERSION"
SPC_CLI_URL="https://github.com/semaphoreci/spc/releases/download/$SPC_CLI_VERSION"
WHEN_CLI_URL="https://github.com/renderedtext/when/releases/download/$WHEN_CLI_VERSION"
TEST_RESULTS_CLI_URL="https://github.com/semaphoreci/test-results/releases/download/$TEST_RESULTS_CLI_VERSION"

create_tarball() {
  tarball_name=$1
  path=$2

  echo "Creating ${tarball_name}..."
  cd ${path}
  tar -cf ${tarball_name} toolbox

  echo "${tarball_name} created. Contents: "
  tar --list --verbose --file=${tarball_name}
}

include_local_binaries() {
  cd ~/$SEMAPHORE_GIT_DIR/cache-cli

  make build OS=linux
  mv bin/cache /tmp/self-hosted-Linux/toolbox

  make build OS=darwin
  mv bin/cache /tmp/self-hosted-Darwin/toolbox

  cd - > /dev/null
}

include_external_linux_binary() {
  url=$1
  binary_name=$2
  destination_path=$3

  echo "Downloading ${binary_name} for linux..."
  curl -s -L --fail --retry 5 ${url}/${binary_name}_Linux_x86_64.tar.gz -o ${destination_path}/${binary_name}_Linux.tar.gz
  cd ${destination_path} && tar -zxf ${binary_name}_Linux.tar.gz && mv ${binary_name} toolbox/ && cd - > /dev/null
}

include_external_darwin_binary() {
  url=$1
  binary_name=$2
  destination_path=$3

  echo "Downloading ${binary_name} for darwin..."
  curl -s -L --fail --retry 5 ${url}/${binary_name}_Darwin_x86_64.tar.gz -o ${destination_path}/${binary_name}_Darwin.tar.gz
  cd ${destination_path} && tar -zxf ${binary_name}_Darwin.tar.gz && mv ${binary_name} toolbox/ && cd - > /dev/null
}

hosted::create_initial_content() {
  echo "Creating initial content of the release directories for the hosted toolbox..."

  rm -rf /tmp/Linux
  rm -rf /tmp/Darwin

  mkdir -p /tmp/Linux/toolbox
  mkdir -p /tmp/Darwin/toolbox

  cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Linux/toolbox
  cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Darwin/toolbox

  exclusions=(.git .gitignore Makefile release tests cache-cli install-self-hosted-toolbox self-hosted-toolbox)
  for exclusion in "${exclusions[@]}"; do
    rm -rf /tmp/Linux/toolbox/${exclusion}
    rm -rf /tmp/Darwin/toolbox/${exclusion}
  done
}

self_hosted::create_initial_content() {
  echo "Creating initial content of the release directories for the self-hosted toolbox..."

  rm -rf /tmp/self-hosted-Linux
  rm -rf /tmp/self-hosted-Darwin

  mkdir -p /tmp/self-hosted-Linux/toolbox
  mkdir -p /tmp/self-hosted-Darwin/toolbox

  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Linux/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Darwin/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Linux/toolbox/toolbox
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Darwin/toolbox/toolbox
  cp ~/$SEMAPHORE_GIT_DIR/libcheckout /tmp/self-hosted-Linux/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/libcheckout /tmp/self-hosted-Darwin/toolbox/
}

self_hosted::pack() {
  self_hosted::create_initial_content
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Linux
  include_external_darwin_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Darwin
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Linux
  include_external_darwin_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Darwin
  include_local_binaries
}

hosted::pack() {
  hosted::create_initial_content
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/Linux
  include_external_darwin_binary $ARTIFACT_CLI_URL "artifact" /tmp/Darwin
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/Linux
  include_external_darwin_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/Darwin
  include_external_linux_binary $SPC_CLI_URL "spc" /tmp/Linux

  echo "Downloading when CLI..."
  curl -s -L --retry 5 $WHEN_CLI_URL/when -o /tmp/Linux/toolbox/when
  chmod +x /tmp/Linux/toolbox/when
}

create_self_hosted=false
while getopts ":a" option; do
  case "${option}" in
    a)
      create_self_hosted=true
      ;;
    \?)
      echo "Invalid option: -$OPTARG" 1>&2
      exit 1
      ;;
  esac
done

shift $((OPTIND -1))

hosted::pack
create_tarball "linux.tar" /tmp/Linux
create_tarball "darwin.tar" /tmp/Darwin

if [[ $create_self_hosted == "true" ]]; then
  self_hosted::pack
  create_tarball "linux.tar" /tmp/self-hosted-Linux
  create_tarball "darwin.tar" /tmp/self-hosted-Darwin
fi