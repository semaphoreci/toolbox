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

#  Elixir
@test "change elixir to 1.7.4" {
  sem-version elixir 1.7.4
  run elixir --version
  assert_line --partial "Elixir 1.7.4"
}

@test "change elixir to 1.11.4" {
  sem-version elixir 1.11.4
  run elixir --version
  assert_line --partial "Elixir 1.11.4"
}

@test "change elixir to 1.12.3" {
  sem-version elixir 1.12.3
  run elixir --version
  assert_line --partial "Elixir 1.12.3"
}

@test "change elixir to 1.13.4" {
  sem-version elixir 1.13.4
  run elixir --version
  assert_line --partial "Elixir 1.13.4"
}
