#!/bin/bash

prefix_cmd() {
  local cmd=$@
  if [ `whoami` == 'root' ]; then
    `$@`
  else
    `sudo $@`
  fi
}

# Before running this, you need to run release/create.sh

# Remove installed toolbox
prefix_cmd rm -rf ~/.toolbox
prefix_cmd rm -f $(which artifact)
prefix_cmd rm -f $(which spc)
prefix_cmd rm -f $(which when)
prefix_cmd rm -f $(which test-results)
prefix_cmd rm -f $(which enetwork)
cd ~
arch=""
case $(uname) in
  Darwin)
    [[ "$(uname -m)" =~ "arm64" ]] && arch="-arm"
    tar -xvf /tmp/Darwin${arch}/darwin${arch}.tar -C /tmp
    mv /tmp/toolbox ~/.toolbox
    ;;
  Linux)
    [[ "$(uname -m)" =~ "aarch" ]] && arch="-arm"
    tar -xvf /tmp/"Linux${arch}"/"linux${arch}".tar -C /tmp
    mv /tmp/toolbox ~/.toolbox
    ;;
esac

cd -

bash ~/.toolbox/install-toolbox
source ~/.toolbox/toolbox
