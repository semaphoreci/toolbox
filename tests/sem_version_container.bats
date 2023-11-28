#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  source ~/.toolbox/toolbox
}

@test "sem-version flutter 3.10.6" {

  run sem-version flutter 3.10.6
  assert_success
  assert_line --partial "3.10.6"
}

@test "sem-version flutter 3.10" {

  run sem-version flutter 3.10
  assert_success
  assert_line --partial "3.10.6"
}

@test "sem-version flutter 3.16.1" {

  run sem-version flutter 3.16.1
  assert_success
  assert_line --partial "3.16.1"
}

@test "sem-version flutter 3.16" {

  run sem-version flutter 3.16
  assert_success
  assert_line --partial "3.16.1"
}
