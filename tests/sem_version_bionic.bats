#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

rm -rf /home/semaphore/.kiex/elixirs
rm -rf /home/semaphore/.kiex/mix/archives
rm -rf /home/semaphore/.kerl/installs/*
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
#  Erlang
@test "change erlang to 21.3" {
  sem-version erlang 21.3
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "21"
}

#  Elixir
@test "change elixir to 1.7.4" {
  sem-version elixir 1.7.4
  assert_success
  run elixir --version
  assert_line --partial "Elixir 1.7.4"
  run ls /home/semaphore/.kiex/mix/archives/elixir-1.7.4/
  assert_success
  assert_line --partial "hex"
}
#  Elixir
@test "change elixir to 1.12.0" {
  sem-version elixir 1.12.0
  assert_success
  run elixir --version
  assert_line --partial "Elixir 1.12.0"
  run ls /home/semaphore/.kiex/mix/archives/elixir-1.12.0/
  assert_success
  assert_line --partial "hex"
}

