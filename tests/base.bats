#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

@test "fail on lftp command not present" {
  sudo apt-get remove -y lftp > /dev/null
  run ./cache usage

  assert_failure
  assert_line --partial "The 'lftp' executable is missing or not in the \$PATH"
  sudo apt-get install -y lftp >/dev/null
}
