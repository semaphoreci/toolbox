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

@test "change java to 25" {
  sem-version java 25
  run java --version
  assert_line --partial "openjdk 25."
}

@test "change java to 17" {
  sem-version java 17
  run java --version
  assert_line --partial "openjdk 17."
}

@test "change java to 21" {
  sem-version java 21
  run java --version
  assert_line --partial "openjdk 21."
}
