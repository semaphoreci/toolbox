#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  cache clear
  rm -rf tmp
}

@test "store file in cache" {
  mkdir tmp && touch tmp/example.file
  run bash -c './cache store --key test-storing --path tmp'

  assert_success
  assert_output --partial "Starting upload"
}

@test "emptying cache directory" {
  mkdir tmp && touch tmp/example.file
  cache store --key test-emptying --path tmp
  run bash -c './cache clear'

  assert_success
  assert_output --partial "Deleting all caches"
}

@test "restoring directory from cache" {
  mkdir tmp && touch tmp/example.file
  cache store --key test-restoring --path tmp/example.file
  rm -rf tmp
  run bash -c './cache restore --key test-restoring'

  assert_success
  assert_output --partial "Transferring from cache repository, using cache key: test-restoring"
  assert_output --partial "Transfer completed"
  assert [ -e "tmp/example.file" ]
}
