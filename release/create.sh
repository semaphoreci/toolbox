#!/bin/bash

set -euo pipefail

ARTIFACT_CLI_VERSION="v0.6.2"
WHEN_CLI_VERSION="v1.0.5"
SPC_CLI_VERSION="v1.9.4"
TEST_RESULTS_CLI_VERSION="v0.6.7"

ARTIFACT_CLI_URL="https://github.com/semaphoreci/artifact/releases/download/$ARTIFACT_CLI_VERSION"
SPC_CLI_URL="https://github.com/semaphoreci/spc/releases/download/$SPC_CLI_VERSION"
TEST_RESULTS_CLI_URL="https://github.com/semaphoreci/test-results/releases/download/$TEST_RESULTS_CLI_VERSION"
WHEN_CLI_URL="https://github.com/renderedtext/when/releases/download/$WHEN_CLI_VERSION"

download_when_cli() {
  rm -rf /tmp/when-cli
  mkdir -p /tmp/when-cli
  echo "Downloading when CLI..."
  curl -sL --retry 5 $WHEN_CLI_URL/when -o /tmp/when-cli/when
  chmod +x /tmp/when-cli/when
}

create_tarball() {
  tarball_name=$1
  path=$2

  echo "Creating ${tarball_name}..."
  cd ${path}
  tar -cf ${tarball_name} toolbox

  echo "${tarball_name} created. Contents: "
  tar --list --verbose --file=${tarball_name}
}

create_checksum() {
# path variable includes the target file
  path=$1
  target_name=$2
  chsum=$(sha256sum $path | awk '{print $1}')
  echo "${chsum} ${target_name}"
}

include_external_linux_binary() {
  url=$1
  binary_name=$2
  destination_path=$3
  arch=$4

  echo "Downloading ${binary_name} for ${arch} for linux..."
  curl -s -L --fail --retry 5 ${url}/${binary_name}_Linux_${arch}.tar.gz -o ${destination_path}/${binary_name}_Linux.tar.gz
  cd ${destination_path} && tar -zxf ${binary_name}_Linux.tar.gz && mv ${binary_name} toolbox/ && cd - > /dev/null
}

include_external_darwin_binary() {
  url=$1
  binary_name=$2
  destination_path=$3
  arch=$4

  echo "Downloading ${binary_name} for ${arch} for darwin..."
  curl -s -L --fail --retry 5 ${url}/${binary_name}_Darwin_${arch}.tar.gz -o ${destination_path}/${binary_name}_Darwin.tar.gz
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
  rm -rf /tmp/Linux-arm
  rm -rf /tmp/Darwin

  mkdir -p /tmp/Linux/toolbox
  mkdir -p /tmp/Linux-arm/toolbox
  mkdir -p /tmp/Darwin/toolbox

  cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Linux/toolbox
  cp -R ~/$SEMAPHORE_GIT_DIR/* /tmp/Linux-arm/toolbox
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
    sem-context
    install-self-hosted-toolbox
    install-self-hosted-toolbox.ps1
    Checkout.psm1
    self-hosted-toolbox
  )
  for exclusion in "${exclusions[@]}"; do
    rm -rf /tmp/Linux/toolbox/${exclusion}
    rm -rf /tmp/Linux-arm/toolbox/${exclusion}
    rm -rf /tmp/Darwin/toolbox/${exclusion}
  done
}

self_hosted::create_initial_content() {
  echo "Creating initial content of the release directories for the self-hosted toolbox..."

  rm -rf /tmp/self-hosted-Linux
  rm -rf /tmp/self-hosted-Linux-arm
  rm -rf /tmp/self-hosted-Darwin
  rm -rf /tmp/self-hosted-Darwin-arm
  rm -rf /tmp/self-hosted-Windows

  mkdir -p /tmp/self-hosted-Linux/toolbox
  mkdir -p /tmp/self-hosted-Linux-arm/toolbox
  mkdir -p /tmp/self-hosted-Darwin/toolbox
  mkdir -p /tmp/self-hosted-Darwin-arm/toolbox
  mkdir -p /tmp/self-hosted-Windows/toolbox

  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Linux/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Linux-arm/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Darwin/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox /tmp/self-hosted-Darwin-arm/toolbox/install-toolbox
  cp ~/$SEMAPHORE_GIT_DIR/install-self-hosted-toolbox.ps1 /tmp/self-hosted-Windows/toolbox/install-toolbox.ps1
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Linux/toolbox/toolbox
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Linux-arm/toolbox/toolbox
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Darwin/toolbox/toolbox
  cp ~/$SEMAPHORE_GIT_DIR/self-hosted-toolbox /tmp/self-hosted-Darwin-arm/toolbox/toolbox

  # Linux/Darwin inclusions
  inclusions=(libcheckout libchecksum retry)
  for inclusion in "${inclusions[@]}"; do
    cp ~/$SEMAPHORE_GIT_DIR/${inclusion} /tmp/self-hosted-Linux/toolbox/
    cp ~/$SEMAPHORE_GIT_DIR/${inclusion} /tmp/self-hosted-Linux-arm/toolbox/
    cp ~/$SEMAPHORE_GIT_DIR/${inclusion} /tmp/self-hosted-Darwin/toolbox/
    cp ~/$SEMAPHORE_GIT_DIR/${inclusion} /tmp/self-hosted-Darwin-arm/toolbox/
  done

  # Windows PowerShell module inclusions
  cp ~/$SEMAPHORE_GIT_DIR/Checkout.psm1 /tmp/self-hosted-Windows/toolbox/

  # For the windows toolbox, we put all the binaries in a bin folder.
  # The reason for that is in Windows we don't have a location like /usr/local/bin
  # to use to place all the binaries in. Instead, we add the '$HOME/.toolbox/bin'
  # folder to the user's PATH.
  mkdir -p /tmp/self-hosted-Windows/toolbox/bin
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/windows/cache.exe /tmp/self-hosted-Windows/toolbox/bin/cache.exe
  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/windows/sem-context.exe /tmp/self-hosted-Windows/toolbox/bin/sem-context.exe
}

self_hosted::pack() {
  self_hosted::create_initial_content
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Linux "x86_64"
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Linux-arm "arm64"
  include_external_darwin_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Darwin "x86_64"
  include_external_darwin_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Darwin-arm "arm64"
  include_external_windows_binary $ARTIFACT_CLI_URL "artifact" /tmp/self-hosted-Windows
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Linux "x86_64"
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Linux-arm "arm64"
  include_external_darwin_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Darwin "x86_64"
  include_external_darwin_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Darwin-arm "arm64"
  include_external_windows_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/self-hosted-Windows
  include_external_linux_binary $SPC_CLI_URL "spc" /tmp/self-hosted-Linux "x86_64"
  include_external_linux_binary $SPC_CLI_URL "spc" /tmp/self-hosted-Linux-arm "arm64"
  include_external_darwin_binary $SPC_CLI_URL "spc" /tmp/self-hosted-Darwin "x86_64"
  include_external_darwin_binary $SPC_CLI_URL "spc" /tmp/self-hosted-Darwin-arm "arm64"

  cp /tmp/when-cli/when /tmp/self-hosted-Linux/toolbox/when
  cp /tmp/when-cli/when /tmp/self-hosted-Linux-arm/toolbox/when
  cp /tmp/when-cli/when /tmp/self-hosted-Darwin/toolbox/when
  cp /tmp/when-cli/when /tmp/self-hosted-Darwin-arm/toolbox/when

  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/linux/amd64/cache /tmp/self-hosted-Linux/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/linux/arm64/cache /tmp/self-hosted-Linux-arm/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/darwin/amd64/cache /tmp/self-hosted-Darwin/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/darwin/arm64/cache /tmp/self-hosted-Darwin-arm/toolbox/

  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/linux/amd64/sem-context /tmp/self-hosted-Linux/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/linux/arm64/sem-context /tmp/self-hosted-Linux-arm/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/darwin/amd64/sem-context /tmp/self-hosted-Darwin/toolbox/
  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/darwin/arm64/sem-context /tmp/self-hosted-Darwin-arm/toolbox/
}

hosted::pack() {
  hosted::create_initial_content
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/Linux "x86_64"
  include_external_linux_binary $ARTIFACT_CLI_URL "artifact" /tmp/Linux-arm "arm64"
  include_external_darwin_binary $ARTIFACT_CLI_URL "artifact" /tmp/Darwin "x86_64"
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/Linux "x86_64"
  include_external_linux_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/Linux-arm "arm64"
  include_external_darwin_binary $TEST_RESULTS_CLI_URL "test-results" /tmp/Darwin "x86_64"
  include_external_linux_binary $SPC_CLI_URL "spc" /tmp/Linux "x86_64"
  include_external_linux_binary $SPC_CLI_URL "spc" /tmp/Linux-arm "arm64"
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/linux/amd64/cache /tmp/Linux/toolbox/cache
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/linux/arm64/cache /tmp/Linux-arm/toolbox/cache
  cp ~/$SEMAPHORE_GIT_DIR/cache-cli/bin/darwin/amd64/cache /tmp/Darwin/toolbox/cache
  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/linux/amd64/sem-context /tmp/Linux/toolbox/sem-context
  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/linux/arm64/sem-context /tmp/Linux-arm/toolbox/sem-context
  cp ~/$SEMAPHORE_GIT_DIR/sem-context/bin/darwin/amd64/sem-context /tmp/Darwin/toolbox/sem-context
  cp /tmp/when-cli/when /tmp/Linux/toolbox/when
  cp /tmp/when-cli/when /tmp/Linux-arm/toolbox/when
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

# The when CLI is a escript that runs on the Erlang VM,
# so it doesn't change based on os/arch. We download it just once before anything else.
download_when_cli
touch /tmp/checksums.txt
truncate -s0 /tmp/checksums.txt
hosted::pack
create_tarball "linux.tar" /tmp/Linux
echo "$(create_checksum /tmp/Linux/linux.tar 'linux.tar')" >> /tmp/checksums.txt
create_tarball "darwin.tar" /tmp/Darwin
echo "$(create_checksum /tmp/Darwin/darwin.tar 'darwin.tar')" >> /tmp/checksums.txt
create_tarball "linux-arm.tar" /tmp/Linux-arm
echo "$(create_checksum /tmp/Linux-arm/linux-arm.tar 'linux-arm.tar')" >> /tmp/checksums.txt

if [[ $create_self_hosted == "true" ]]; then
  self_hosted::pack
  create_tarball "linux.tar" /tmp/self-hosted-Linux
  echo "$(create_checksum /tmp/self-hosted-Linux/linux.tar 'self-hosted-linux.tar')" >> /tmp/checksums.txt
  create_tarball "linux-arm.tar" /tmp/self-hosted-Linux-arm
  echo "$(create_checksum /tmp/self-hosted-Linux-arm/linux-arm.tar 'self-hosted-linux-arm.tar')" >> /tmp/checksums.txt
  create_tarball "darwin.tar" /tmp/self-hosted-Darwin
  echo "$(create_checksum /tmp/self-hosted-Darwin/darwin.tar 'self-hosted-darwin.tar')" >> /tmp/checksums.txt
  create_tarball "darwin-arm.tar" /tmp/self-hosted-Darwin-arm
  echo "$(create_checksum /tmp/self-hosted-Darwin-arm/darwin-arm.tar 'self-hosted-darwin-arm.tar')" >> /tmp/checksums.txt
  create_tarball "windows.tar" /tmp/self-hosted-Windows
  echo "$(create_checksum /tmp/self-hosted-Windows/windows.tar 'self-hosted-windows.tar')" >> /tmp/checksums.txt
fi

cat /tmp/checksums.txt
