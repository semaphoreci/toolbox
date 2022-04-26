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


