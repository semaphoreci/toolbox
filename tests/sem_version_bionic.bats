#!/usr/bin/env bats

rm -rf /home/semaphore/.kiex/elixirs/*
rm -rf /home/semaphore/.kiex/mix/achives/*

rm -rf /home/semaphore/.kerl/installs/*
rm -rf /home/semaphore/.kerl/otp_installations

load "support/bats-support/load"
load "support/bats-assert/load"

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
#  Firefox

@test "change firefox to 52" {

  run sem-version firefox 52
  assert_success
  assert_line --partial "Mozilla Firefox 52"
}
@test "change firefox to 78" {

  run sem-version firefox 78
  assert_success
  assert_line --partial "Mozilla Firefox 78"
}
@test "change firefox to 90" {

  run sem-version firefox 90
  assert_failure
}

#  Ruby
@test "change ruby to 2.5.3" {

  run sem-version ruby 2.5.3
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.5.3"
}

@test "change ruby to 2.3.8" {

  run sem-version ruby 2.3.8
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.3.8"
}

@test "change ruby to 2.7.4" {

  run sem-version ruby 2.7.4
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.7.4"
}

@test "change ruby to 3.0.2" {

  run sem-version ruby 3.0.2
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.0.2"
}
@test "ruby minor versions test" {

  run sem-version ruby 2.5
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.5.9"

  run sem-version ruby 2.6
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.6.8"

  run sem-version ruby 2.7
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.7.4"

  run sem-version ruby 3.0
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.0.2"

}


@test "change ruby to 4.0.1" {

  run sem-version ruby 4.0.1
  assert_failure
}

#  C
@test "change gcc to 8" {

  run sem-version c 8
  assert_success
  run gcc -v
  assert_line --partial "gcc version 8."
}
@test "change gcc to 16" {

  run sem-version c 16
  assert_failure
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
@test "change php to 7.3.29" {

  run sem-version php 7.3.29
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 7.3.29"
  run php -m
  assert_line --partial "magick"
  assert_line --partial "gd"
  assert_line --partial "imap"
}
@test "change php to 7.4.21" {

  run sem-version php 7.4.21
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 7.4.21"
  run php -m 
  assert_line --partial "magick"
  assert_line --partial "gd"
  assert_line --partial "imap"
}
@test "change php to 8.0.8" {

  run sem-version php 8.0.8
  assert_success
  source ~/.phpbrew/bashrc
  run php -v
  assert_line --partial "PHP 8.0.8"
  run php -m 
  assert_line --partial "gd"
  assert_line --partial "imap"
}
@test "php check composer 8.0.8" {

  run which composer
  assert_success
  source ~/.phpbrew/bashrc
  assert_line --partial "8.0.8"
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
#  Node
@test "change node to 12.16.1" {
  sem-version node 12.16.1
  run node --version
  assert_line --partial "v12.16.1"
}
#  kubectl
@test "change kubectl to 1.15.3" {
  sem-version kubectl 1.15.3
  run kubectl version
  assert_line --partial "1.15.3"
}
#  erlang
@test "change erlang to 23.2" {
  sem-version erlang 23.2
  run erl -eval 'erlang:display(erlang:system_info(otp_release)), halt().'  -noshell
  assert_line --partial "23"
}
#  scala
@test "change scala to 2.11" {

  run scala -version
  assert_line --partial "2.12"

  sem-version scala 2.11
  run scala -version
  assert_line --partial "2.11"
}

@test "sem-version fail php" {

  run sem-version php 9
  assert_failure
}

@test "sem-version firefox 90" {

  run sem-version firefox 90
  assert_failure
}

@test "sem-version test ignore" {

  run sem-version firefox 90
  assert_failure
  run sem-version firefox 90 --ignore
  assert_success
}

@test "sem-version go path check" {

  sem-version go 1.16
  run echo ${PATH}
  assert_line --partial "$(go env GOPATH)/bin"
}


