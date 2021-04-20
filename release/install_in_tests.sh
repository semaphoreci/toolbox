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

cd ~

case $(uname) in
  Darwin)
    tar -xvf /tmp/Darwin/darwin.tar -C /tmp
    mv /tmp/toolbox ~/.toolbox
    ;;
  Linux)
    tar -xvf /tmp/Linux/linux.tar -C /tmp
    mv /tmp/toolbox ~/.toolbox
    ;;
esac

cd -

bash ~/.toolbox/install-toolbox
source ~/.toolbox/toolbox
