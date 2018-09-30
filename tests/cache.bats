#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  cache clear
  rm -rf tmp
}

################################################################################
# cache store
################################################################################

@test "store local file to cache repository" {
  mkdir tmp && touch tmp/example.file
  run bash -c './cache store --key test-storing --path tmp'

  assert_success
  assert_output --partial "Uploading 'tmp' with cache key 'test-storing'"
  assert_output --partial "Upload complete."
}

@test "store nonexistent local file to cache repository" {
  run bash -c './cache store --key test-storing --path tmp'

  assert_success
  assert_output --partial "Starting upload"
  assert_output --partial "'tmp' doesn't exist locally, skipping."
}

@test "store with key which is already present in cache repository" {
  mkdir tmp && touch tmp/example.file
  cache store --key test-storing --path tmp
  ./cache list
  run bash -c './cache store --key test-storing --path tmp'

  assert_success
  assert_output --partial "Key 'test-storing' already present on remote."
  assert_output --partial "Upload complete."
}

################################################################################
# cache restore
################################################################################

@test "restoring existing directory from cache and perserving the directory hierarchy" {
  mkdir tmp && mkdir tmp/first && mkdir tmp/first/second && touch tmp/first/second/example.file
  cache store --key restore-dir-hierarchy --path tmp/first/second
  rm -rf tmp
  run ./cache restore --key restore-dir-hierarchy

  assert_success
  assert [ -e "tmp/first/second/example.file" ]
  assert_output --partial "Using cache key: restore-dir-hierarchy."
  assert_output --partial "Transferring from cache repository, using cache key: restore-dir-hierarchy."
  assert_output --partial "Transfer completed."
}

@test "restoring nonexistent directory from cache" {
  run bash -c './cache restore --key test'

  assert_success
  assert_output --partial "Using cache key: test".
  assert_output --partial "Key 'test' does not exist on cache repository, skipping restore."
}

################################################################################
# cache clear
################################################################################

@test "emptying cache repository when cache is not empty" {
  mkdir tmp && touch tmp/example.file
  cache store --key test-emptying --path tmp
  run bash -c './cache clear'

  assert_success
  assert_output --partial "Deleting all caches"
}

@test "emptying cache repository when cache is empty" {
  run bash -c './cache clear'

  assert_success
}

################################################################################
# cache list
################################################################################

@test "listing cache repository when it has cache keys" {
  mkdir tmp && touch tmp/example.file
  cache store --key listing-v1 --path tmp
  cache store --key listing-v2 --path tmp
  run bash -c './cache list'

  assert_success
  assert_output --partial "Listing available keys in cache repository"
  assert_output --partial "listing-v1"
  assert_output --partial "listing-v2"
  assert_output --partial "Listed available keys in cache repository"
}

@test "listing cache keys when cache is empty" {
  run bash -c './cache clear'

  assert_success
}
