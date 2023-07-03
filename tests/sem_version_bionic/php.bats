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

# PHP
@test "change php to 7.2.34" {

  run sem-version php 7.2.34
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 7.2.34"
  run php -m
  assert_line --partial "magick"
  assert_line --partial "gd"
  assert_line --partial "imap"
}

@test "change php to 7.3.33" {

  run sem-version php 7.3.33
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 7.3.33"
  run php -m
  assert_line --partial "magick"
  assert_line --partial "gd"
  assert_line --partial "imap"
}

@test "change php to 7.4.33" {

  run sem-version php 7.4.33
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 7.4.33"
  run php -m 
  assert_line --partial "magick"
  assert_line --partial "gd"
  assert_line --partial "imap"
}

@test "change php to 8.0.29" {

  run sem-version php 8.0.29
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 8.0.29"
  run php -m 
  assert_line --partial "gd"
  assert_line --partial "imap"
}

@test "php check composer 8.0.29" {

  run which composer
  assert_success
  source ~/.phpbrew/bashrc
  assert_line --partial "8.0.29"
}

@test "php check source 8.0.29" {

  run sem-version php 8.0.29
  assert_success
  source ~/.phpbrew/bashrc
  assert_line --partial "8.0.29"
  run phpbrew ext install iconv
  assert_success
}
