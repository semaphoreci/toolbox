#!/usr/bin/env bats

load "../support/bats-support/load"
load "../support/bats-assert/load"

setup() {
  source /tmp/.env
  source /opt/change-erlang-version.sh
  source /opt/change-python-version.sh
  source /opt/change-go-version.sh
  source /opt/change-java-version.sh
  source /opt/change-scala-version.sh
  source /opt/change-firefox-version.sh
  source ~/.phpbrew/bashrc
  . /home/semaphore/.nvm/nvm.sh
  export PATH="$PATH:/home/semaphore/.yarn/bin"
  source "/home/semaphore/.kiex/scripts/kiex"
  export PATH="/home/semaphore/.rbenv/bin:$PATH"
  export NVM_DIR=/home/semaphore/.nvm
  export PHPBREW_HOME=/home/semaphore/.phpbrew
  eval "$(rbenv init -)"

  source ~/.toolbox/toolbox
}

#  Node
@test "change node to 12.16.1" {
  sem-version node 12.16.1
  run node --version
  assert_line --partial "v12.16.1"
}

@test "change node to 16.15.1" {
  sem-version node 16.15.1
  run node --version
  assert_line --partial "v16.15.1"
}
