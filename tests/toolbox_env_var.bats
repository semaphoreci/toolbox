#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  TMPFILE=$(mktemp /tmp/toolbox-XXXX)
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
  export SEMAPHORE_TOOLBOX_PACKAGE_HOST="packages2.semaphoreci.com"
  export SEMAPHORE_TOOLBOX_SERVICE_HOST="packages2.semaphoreci.com"
  source ~/.toolbox/toolbox
}

@test "package url should contain packages2" {
  run install::package_url
  assert_output --partial "packages2"
}

@test "package url shouldn't contain packages2" {
  
  run unset SEMAPHORE_TOOLBOX_PACKAGE_HOST && install::package_url
  refute_output --partial "packages2"
}

@test "sem-service start postgres 16 should fail" {
  sem-service start postgres 16 2>&1 > /dev/null || true
  run psql -h 0.0.0.0 -U postgres -c "SELECT version()"
  assert_failure
}

@test "sem-service start postgres 16 should succeed" {
  unset SEMAPHORE_TOOLBOX_SERVICE_HOST
  sem-service start postgres 16
  sleep 20
  run psql -h 0.0.0.0 -U postgres -c "SELECT version()"
  assert_output --partial "16"
  assert_success
}




