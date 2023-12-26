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
  source "/home/semaphore/.kiex/scripts/kiex"
  export PATH="/home/semaphore/.rbenv/bin:$PATH"
  export NVM_DIR=/home/semaphore/.nvm
  export PHPBREW_HOME=/home/semaphore/.phpbrew
  eval "$(rbenv init -)"

  source ~/.toolbox/toolbox
}

#  Elixir
@test "change elixir to 1.14.0" {
  sem-version elixir 1.14.0
  run elixir --version
  assert_line --partial "Elixir 1.14.0"
}

@test "change elixir to 1.14.5" {
  sem-version elixir 1.14.5
  run elixir --version
  assert_line --partial "Elixir 1.14.5"
}

@test "change elixir to 1.15.0" {
  sem-version elixir 1.15.0
  run elixir --version
  assert_line --partial "Elixir 1.15.0"
}

@test "change elixir to 1.15.7" {
  sem-version elixir 1.15.7
  run elixir --version
  assert_line --partial "Elixir 1.15.7"
}

@test "change elixir to 1.16.0" {
  sem-version elixir 1.16.0
  run elixir --version
  assert_line --partial "Elixir 1.16.0"
}
