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

#  Elixir
@test "change elixir to 1.12" {
  sem-version elixir 1.12
  run elixir --version
  assert_line --partial "Elixir 1.12.3"
}

@test "change elixir to 1.13" {
  sem-version elixir 1.13
  run elixir --version
  assert_line --partial "Elixir 1.13.4"
}

@test "change elixir to 1.14" {
  sem-version elixir 1.14
  run elixir --version
  assert_line --partial "Elixir 1.14.5"
}

@test "change elixir to 1.15" {
  sem-version elixir 1.15
  run elixir --version
  assert_line --partial "Elixir 1.15.8"
}

@test "change elixir to 1.16" {
  sem-version elixir 1.16
  run elixir --version
  assert_line --partial "Elixir 1.16.3"
}

@test "change elixir to 1.17" {
  sem-version elixir 1.17
  run elixir --version
  assert_line --partial "Elixir 1.17.3"
}

@test "change elixir to 1.18" {
  sem-version elixir 1.18
  run elixir --version
  assert_line --partial "Elixir 1.18.4"
}

@test "change elixir to 1.19" {
  sem-version elixir 1.19
  run elixir --version
  assert_line --partial "Elixir 1.19.5"
}

@test "change elixir to 1.20" {
  sem-version elixir 1.20
  run elixir --version
  assert_line --partial "Elixir 1.20.1"
}
