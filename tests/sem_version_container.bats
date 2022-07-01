#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  source ~/.toolbox/toolbox
}

@test "sem-version flutter 2.10.5" {

  run sem-version flutter 2.10.5
  assert_success
  assert_line --partial "2.10.5"
}


@test "sem-version flutter 2.10" {

  run sem-version flutter 2.10
  assert_success
  assert_line --partial "2.10"
}


@test "sem-version flutter 3.0.1" {

  run sem-version flutter 3.0.1
  assert_success
  assert_line --partial "3.0.1"
}

@test "sem-version flutter 3.0" {

  run sem-version flutter 3.0
  assert_success
  assert_line --partial "3.0.1"
}
