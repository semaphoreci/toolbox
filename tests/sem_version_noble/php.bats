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

# PHP
@test "change php to 8.1" {

  run sem-version php 8.1
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 8.1.29"
  run php -m
  assert_line --partial "gd"
  assert_line --partial "imap"
  run which composer
  assert_success
  assert_line --partial "8.1.29"
  run phpbrew ext install iconv
  assert_success
}

@test "change php to 8.2" {

  run sem-version php 8.2
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 8.2.31"
  run php -m
  assert_line --partial "gd"
  assert_line --partial "imap"
  run which composer
  assert_success
  assert_line --partial "8.2.31"
  run phpbrew ext install iconv
  assert_success
}

@test "change php to 8.3" {

  run sem-version php 8.3
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 8.3.31"
  run php -m
  assert_line --partial "gd"
  assert_line --partial "imap"
  run which composer
  assert_success
  assert_line --partial "8.3.31"
  run phpbrew ext install iconv
  assert_success
}

@test "change php to 8.4" {

  run sem-version php 8.4
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 8.4.22"
  run php -m
  assert_line --partial "gd"
  run which composer
  assert_success
  assert_line --partial "8.4.22"
  run phpbrew ext install iconv
  assert_success
}
