#!/bin/bash

# Before running this, you need to run release/create.sh

# Remove installed toolbox
rm -rf ~/.toolbox
rm -f $(whereis artifact)
rm -f $(whereis spc)
rm -f $(whereis when)

cd ~

case $(uname) in
  Darwin) tar -xvf /tmp/Darwin/darwin.tar ;;
  Linux)  tar -xvf /tmp/Linux/linux.tar   ;;
esac

bash ~/.toolbox/install-toolbox
source ~/.toolbox/toolbox
