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
  source ~/.phpbrew/bashrc
  . /home/semaphore/.nvm/nvm.sh
  export PATH="$PATH:/home/semaphore/.yarn/bin"
  export KIEX_HOME="$HOME/.kiex"
  source "/home/semaphore/.kiex/scripts/kiex"
  export PATH="/home/semaphore/.rbenv/bin:$PATH"
  export NVM_DIR=/home/semaphore/.nvm
  export PHPBREW_HOME=/home/semaphore/.phpbrew
  eval "$(rbenv init -)"

  source ~/.toolbox/toolbox
}

#  kubectl
@test "change kubectl to 1.15.3" {
  sem-version kubectl 1.15.3
  run kubectl version
  assert_line --partial "1.15.3"
}

@test "change kubectl to 1.28.13" {
  sem-version kubectl 1.28.13
  run kubectl version
  assert_line --partial "1.28.13"
}

@test "change kubectl to 1.29.8" {
  sem-version kubectl 1.29.8
  run kubectl version
  assert_line --partial "1.29.8"
}

@test "change kubectl to 1.30.4" {
  sem-version kubectl 1.30.4
  run kubectl version
  assert_line --partial "1.30.4"
}

@test "change kubectl to 1.31.0" {
  sem-version kubectl 1.31.0
  run kubectl version
  assert_line --partial "1.31.0"
}
