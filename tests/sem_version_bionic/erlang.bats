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

#  erlang
@test "change erlang to 23.3" {
  sem-version erlang 23.3
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "23"
}

@test "change erlang to 24.0" {
  sem-version erlang 24.0
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "24"
}

@test "change erlang to 24.3" {
  sem-version erlang 24.3
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "24"
}

@test "change erlang to 25.0" {
  sem-version erlang 25.0
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "25"
}

@test "change erlang to 25.1" {
  sem-version erlang 25.1
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "25"
}

@test "change erlang to 25.2" {
  sem-version erlang 25.2
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "25"
}

@test "change erlang to 25.3" {
  sem-version erlang 25.3
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "25"
}

@test "change erlang to 26.0" {
  sem-version erlang 26.0
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "26"
}
