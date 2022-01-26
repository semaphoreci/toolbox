#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

@test "Account settings missing" {

  run enetwork start
  assert_success
  assert_line --partial "Authentication error, not setting docker gateway"
}
