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

@test "sem-version go 1.18.10" {

  sem-version go 1.18.10
  run echo ${PATH}
  assert_line --partial "$(go env GOPATH)/bin"
  run go version
  assert_line --partial "go1.18.10"
}

@test "sem-version go 1.19.9" {

  sem-version go 1.19.9
  run echo ${PATH}
  assert_line --partial "$(go env GOPATH)/bin"
  run go version
  assert_line --partial "go1.19.9"
}

@test "sem-version go 1.20.4" {

  sem-version go 1.20.4
  run echo ${PATH}
  assert_line --partial "$(go env GOPATH)/bin"
  run go version
  assert_line --partial "go1.20.4"
}

@test "sem-version go 1.21.1" {

  sem-version go 1.21.1
  run echo ${PATH}
  assert_line --partial "$(go env GOPATH)/bin"
  run go version
  assert_line --partial "go1.21.1"
}

@test "sem-version go 1.24.11" {

  sem-version go 1.24.11
  run echo ${PATH}
  assert_line --partial "$(go env GOPATH)/bin"
  run go version
  assert_line --partial "go1.24.11"
}

@test "sem-version go 1.25.5" {

  sem-version go 1.25.5
  run echo ${PATH}
  assert_line --partial "$(go env GOPATH)/bin"
  run go version
  assert_line --partial "go1.25.5"
}
