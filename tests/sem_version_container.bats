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

@test "sem-version flutter 2.5.1" {

  run sem-version flutter 2.5.1
  assert_success
  assert_line --partial "2.5.1"
}


