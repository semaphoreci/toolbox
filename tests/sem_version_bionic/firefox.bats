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

#  Firefox

@test "change firefox to 68" {
 
  run sem-version firefox 68
  assert_success
  assert_line --partial "Mozilla Firefox 68"
}

@test "change firefox to 78" {

  run sem-version firefox 78
  assert_success
  assert_line --partial "Mozilla Firefox 78"
}

@test "change firefox to 102" {

  run sem-version firefox 102
  assert_success
  assert_line --partial "Mozilla Firefox 102"
}
