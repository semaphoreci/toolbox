#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

@test "store file in cache" {
  mkdir tmp && touch tmp/example.file
  run bash -c './cache store --key v4 --path tmp'

  assert_success
  assert_output --partial "Starting upload"
}
