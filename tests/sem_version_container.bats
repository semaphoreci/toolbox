#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  source ~/.toolbox/toolbox
}

@test "sem-version flutter 2.5.3" {

  run sem-version flutter 2.5.3
  assert_success
  assert_line --partial "2.5.3"
}

@test "sem-version flutter 2.8.1" {

  run sem-version flutter 2.8.1
  assert_success
  assert_line --partial "2.8.1"
}

@test "sem-version flutter 2.8" {

  run sem-version flutter 2.8
  assert_success
  assert_line --partial "2.8.1"
}


