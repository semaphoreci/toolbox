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
@test "change ruby to 2.7.8" {

  run sem-version ruby 2.7.8
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.7.8"
}

@test "change ruby to 3.0.6" {

  run sem-version ruby 3.0.6
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.0.6"
}

@test "change ruby to 3.1.7" {

  run sem-version ruby 3.1.7
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.1.7"
}

@test "change ruby to 3.2.9" {

  run sem-version ruby 3.2.9
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.2.9"
}

@test "change ruby to 3.3.10" {

  run sem-version ruby 3.3.10
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.3.10"
}

@test "change ruby to 3.4.8" {

  run sem-version ruby 3.4.8
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.4.8"
}

@test "change ruby to 4.0.0" {

  run sem-version ruby 4.0.0
  assert_success
  run ruby --version
  assert_line --partial "ruby 4.0.0"
}

@test "ruby minor versions test" {

  run sem-version ruby 2.7
  assert_success
  run ruby --version
  assert_line --partial "ruby 2.7.8"

  run sem-version ruby 3.0
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.0.7"

  run sem-version ruby 3.1
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.1.7"

  run sem-version ruby 3.2
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.2.9"

  run sem-version ruby 3.3
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.3.10"

  run sem-version ruby 3.4
  assert_success
  run ruby --version
  assert_line --partial "ruby 3.4.8"

  run sem-version ruby 4.0
  assert_success
  run ruby --version
  assert_line --partial "ruby 4.0.0"
}

@test "change ruby to 5.0.1" {

  run sem-version ruby 5.0.1
  assert_failure
}
