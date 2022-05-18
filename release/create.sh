#!/bin/bash

set -euo pipefail

ARTIFACT_CLI_VERSION="v0.5.5"
WHEN_CLI_VERSION="v1.0.5"
SPC_CLI_VERSION="v1.9.2"
TEST_RESULTS_CLI_VERSION="v0.4.12"

ARTIFACT_CLI_URL="https://github.com/semaphoreci/artifact/releases/download/$ARTIFACT_CLI_VERSION"
SPC_CLI_URL="https://github.com/semaphoreci/spc/releases/download/$SPC_CLI_VERSION"
TEST_RESULTS_CLI_URL="https://github.com/semaphoreci/test-results/releases/download/$TEST_RESULTS_CLI_VERSION"
WHEN_CLI_URL="https://github.com/renderedtext/when/releases/download/$WHEN_CLI_VERSION"

create_tarball() {
  tarball_name=$1
  path=$2

  echo "Creating ${tarball_name}..."
  cd ${path}
  tar -cf ${tarball_name} toolbox

  echo "${tarball_name} created. Contents: "
  tar --list --verbose --file=${tarball_name}
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

include_external_windows_binary() {
  url=$1
  binary_name=$2
  destination_path=$3

  echo "Downloading ${binary_name} for windows..."
  curl -s -L --fail --retry 5 ${url}/${binary_name}_Windows_x86_64.tar.gz -o ${destination_path}/${binary_name}_Windows.tar.gz
  cd ${destination_path} && tar -zxf ${binary_name}_Windows.tar.gz && mv ${binary_name}.exe toolbox/bin/ && cd - > /dev/null
}

hosted::create_initial_content() {
  echo "Creating initial content of the release directories for the hosted toolbox..."

  rm -rf /tmp/Linux
  rm -rf /tmp/Darwin

  mkdir -p /tmp/Linux/toolbox
  mkdir -p /tmp/Darwin/toolbox

  cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Linux/toolbox
  cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Darwin/toolbox

  exclusions=(
    .git
    .gitignore
    Makefile
    release
    scripts
    tests
    docker-compose.yml
    cache-cli
    install-self-hosted-toolbox
    install-self-hosted-toolbox.ps1
    Checkout.psm1
    self-hosted-toolbox
  )
  for exclusion in "${exclusions[@]}"; do
    rm -rf /tmp/Linux/toolbox/${exclusion}
    rm -rf /tmp/Darwin/toolbox/${exclusion}
  done
}

self_hosted::create_initial_content() {
  echo "Creating initial content of the release directories for the self-hosted toolbox..."

  rm -rf /tmp/self-hosted-Linux
  rm -rf /tmp/self-hosted-Darwin
  rm -rf /tmp/self-hosted-Windows

  mkdir -p /tmp/self-hosted-Linux/toolbox
  mkdir -p /tmp/self-hosted-Darwin/toolbox
  mkdir -p /tmp/self-hosted-Windows/toolbox

  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Linux/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Darwin/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox.ps1 /tmp/self-hosted-Windows/toolbox/install-toolbox.ps1
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Linux/toolbox/toolbox
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Darwin/toolbox/toolbox

  # Linux/Darwin inclusions
  inclusions=(libcheckout libchecksum retry)
  for inclusion in "${inclusions[@]}"; do
    cp ~/$SEMAPHORE_GIT_DIR/${inclusion} /tmp/self-hosted-Linux/toolbox/
    cp ~/$SEMAPHORE_GIT_DIR/${inclusion} /tmp/self-hosted-Darwin/toolbox/
  done

  # Windows PowerShell module inclusions
  cp ~/$SEMAPHORE_GIT_DIR/Checkout.psm1 /tmp/self-hosted-Windows/toolbox/

  # For the windows toolbox, we put all the binaries in a bin folder.
  # The reason for that is in Windows we don't have a location like /usr/local/bin
  # to use to place all the binaries in. Instead, we add the '$HOME/.toolbox/bin'
  # folder to the user's PATH.
  mkdir -p /tmp/self-hosted-Windows/toolbox/bin
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/windows/cache.exe /tmp/self-hosted-Windows/toolbox/bin/cache.exe
}

self_hosted::pack() {
  self_hosted::create_initial_content
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Linux
  include_external_darwin_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Darwin
  include_external_windows_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Windows
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Linux
  include_external_darwin_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Darwin
  include_external_linux_binary $SPC_CLI_URL "spc" /tmp/self-hosted-Linux
  include_external_darwin_binary $SPC_CLI_URL "spc" /tmp/self-hosted-Darwin

  curl -s -L --retry 5 $WHEN_CLI_URL/when -o /tmp/self-hosted-Linux/toolbox/when
  chmod +x /tmp/self-hosted-Linux/toolbox/when
  curl -s -L --retry 5 $WHEN_CLI_URL/when -o /tmp/self-hosted-Darwin/toolbox/when
  chmod +x /tmp/self-hosted-Darwin/toolbox/when

  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/linux/cache /tmp/self-hosted-Linux/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/darwin/cache /tmp/self-hosted-Darwin/toolbox/
}

hosted::pack() {
  hosted::create_initial_content
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/Linux
  include_external_darwin_binary $ARTIFACT_CLI_URL "artifact" /tmp/Darwin
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/Linux
  include_external_darwin_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/Darwin
  include_external_linux_binary $SPC_CLI_URL "spc" /tmp/Linux
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/linux/cache /tmp/Linux/toolbox/new-cache
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/darwin/cache /tmp/Darwin/toolbox/new-cache

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
  create_tarball "windows.tar" /tmp/self-hosted-Windows
fi
