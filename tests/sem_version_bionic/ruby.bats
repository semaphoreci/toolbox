#!/usr/bin/env bats

load "../support/bats-support/load"
load "../support/bats-assert/load"

setup() {
  source /tmp/.env-*
  export PATH="/home/semaphore/.rbenv/bin:$PATH"
  eval "$(rbenv init -)"
  set -u
  source ~/.toolbox/toolbox
}

#  Ruby
@test "change ruby to 2.3.8" {

  run sem-version ruby 2.3.8
  assert_success    
  run ruby --version
  assert_line --partial "ruby 2.3.8"
}

@test "change ruby to 2.5.3" {

  run sem-version ruby 2.5.3
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.5.3"
}

@test "change ruby to 2.6.10" {

  run sem-version ruby 2.6.10
  assert_success    
  run ruby --version
  assert_line --partial "ruby 2.6.10"
}

@test "change ruby to 2.7.6" {

  run sem-version ruby 2.7.6
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.7.6"
}

@test "change ruby to 3.0.4" {

  run sem-version ruby 3.0.4
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.0.4"
}

@test "change ruby to 3.1.0" {

  run sem-version ruby 3.1.0
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.1.0"
}

@test "ruby minor versions test" {

  run sem-version ruby 2.5
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.5.9"

  run sem-version ruby 2.6
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.6.10"

  run sem-version ruby 2.7
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.7.6"

  run sem-version ruby 3.0
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.0.4"

  run sem-version ruby 3.1
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.1.2"
}

@test "change ruby to 4.0.1" {

  run sem-version ruby 4.0.1
  assert_failure
}
