#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

@test "artifacts - uploading to proect level" {
  run artifact push --force project ~/.toolbox/retry

  assert_success
}

@test "artifacts - uploading to workflows level" {
  run artifact push --force workflows ~/.toolbox/retry

  assert_success
}

@test "artifacts - uploading to job level" {
  run artifact push --force job ~/.toolbox/retry

  assert_success
}
