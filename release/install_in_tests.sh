#!/bin/bash

# Before running this, you need to run release/create.sh

# Remove installed toolbox
sudo rm -rf ~/.toolbox
sudo rm -f $(which artifact)
sudo rm -f $(which spc)
sudo rm -f $(which when)

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

bash ~/.toolbox/install-toolbox
cat ~/.toolbox/toolbox >> ~/.bash_profile
source ~/.bash_profile
