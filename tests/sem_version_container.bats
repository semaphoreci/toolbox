#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  source ~/.toolbox/toolbox
}

@test "sem-version flutter 2.2.3" {

  run sem-version flutter 2.2.3
  assert_success
  assert_line --partial "2.2.3"
}

@test "sem-version flutter 2.5.2" {

  run sem-version flutter 2.5.2
  assert_success
  assert_line --partial "2.5.2"
}

@test "sem-version flutter 2.5" {

  run sem-version flutter 2.5
  assert_success
  assert_line --partial "2.5.2"
}


