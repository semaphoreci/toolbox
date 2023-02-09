#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  source ~/.toolbox/toolbox
}

@test "sem-version flutter 3.0.5" {

  run sem-version flutter 3.0.5
  assert_success
  assert_line --partial "3.0.5"
}

@test "sem-version flutter 3.0" {

  run sem-version flutter 3.0
  assert_success
  assert_line --partial "3.0.5"
}

@test "sem-version flutter 3.3.0" {

  run sem-version flutter 3.3.0
  assert_success
  assert_line --partial "3.3.0"
}


@test "sem-version flutter 3.3" {

  run sem-version flutter 3.3
  assert_success
  assert_line --partial "3.3.0"
}