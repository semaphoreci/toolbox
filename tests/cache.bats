#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  cache clear
  rm -rf tmp
}

@test "store existing file in cache" {
  mkdir tmp && touch tmp/example.file
  run bash -c './cache store --key test-storing --path tmp'

  assert_success
  assert_output --partial "Uploading 'tmp' with cache key 'test-storing'"
  assert_output --partial "Upload complete."
}

@test "store nonexistent file in cache" {
  run bash -c './cache store --key test-storing --path tmp'

  assert_success
  assert_output --partial "Starting upload"
  assert_output --partial "'tmp' doesn't exist locally, skipping."
}

@test "emptying cache directory when cache is not empty" {
  mkdir tmp && touch tmp/example.file
  cache store --key test-emptying --path tmp
  run bash -c './cache clear'

  assert_success
  assert_output --partial "Deleting all caches"
}

@test "emptying cache directory when cache is empty" {
  run bash -c './cache clear'

  assert_success
}

@test "restoring existing directory from cache" {
  mkdir tmp && touch tmp/example.file
  cache store --key test-restoring --path tmp/example.file
  rm -rf tmp
  run bash -c './cache restore --key test-restoring'

  assert_success
  assert_output --partial "Transferring from cache repository, using cache key: test-restoring"
  assert_output --partial "Transfer completed"
  assert [ -e "tmp/example.file" ]
}

@test "restoring nonexistent directory from cache" {
  run bash -c './cache restore --key test'

  assert_success
  assert_output --partial "Transferring from cache repository, using cache key: test"
  assert_output --partial "'test' does not exist on cache repository, skipping restore."
}
