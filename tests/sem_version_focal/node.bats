#!/usr/bin/env bats

load "../support/bats-support/load"
load "../support/bats-assert/load"

setup() {
  source /tmp/.env-*
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

@test "change node to 18.20.8" {
  sem-version node 18.20.8
  run node --version
  assert_line --partial "v18.20.8"
}

@test "change node to 22.21.1" {
  sem-version node 22.21.1
  run node --version
  assert_line --partial "v22.21.1"
}

@test "change node to 23.11.1" {
  sem-version node 23.11.1
  run node --version
  assert_line --partial "v23.11.1"
}

@test "change node to 24.12.0" {
  sem-version node 24.12.0
  run node --version
  assert_line --partial "v24.12.0"
}

@test "change node to 25.2.1" {
  sem-version node 25.2.1
  run node --version
  assert_line --partial "v25.2.1"
}

@test "change node to 30.30.30" {
  run sem-version node 30.30.30
  assert_failure
}
