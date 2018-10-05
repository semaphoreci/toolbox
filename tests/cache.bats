#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  ./cache clear
  rm -rf tmp
}

################################################################################
# cache --verbose
################################################################################

@test "verbose flag logs detailed steps" {
  skip "option is not public yet"
  run ./cache --verbose

  assert_success
  assert_output --partial "Checking environment variables"
}

@test "default logs without verbose output" {
  run ./cache

  assert_success
  refute_output --partial "Checking if LFPT is present"
  refute_output --partial "Loading SSH key into the agent"
  refute_output --partial "Checking environment variables"
}
################################################################################
# cache store
################################################################################

@test "save local file to cache store" {
  mkdir tmp && touch tmp/example.file
  run ./cache store --key test-storing --path tmp

  assert_success
  assert_output --partial "Uploading 'tmp' with cache key 'test-storing'"
  assert_output --partial "Upload complete."

  run ./cache has_key test-storing
  assert_success
}

@test "save nonexistent local file to cache store" {
  run ./cache store --key test-storing --path tmp

  assert_success
  assert_output --partial "Starting upload"
  assert_output --partial "'tmp' doesn't exist locally, skipping."
}

@test "store with key which is already present in cache" {
  mkdir tmp && touch tmp/example.file
  ./cache store --key test-storing --path tmp

  run ./cache has_key test-storing
  assert_success

  run ./cache store --key test-storing --path tmp

  assert_success
  assert_output --partial "Key 'test-storing' already present on remote."
  assert_output --partial "Upload complete."

  run ./cache has_key test-storing
  assert_success
}

################################################################################
# cache restore
################################################################################

@test "restoring existing directory from cache and perserving the directory hierarchy" {
  mkdir tmp && mkdir tmp/first && mkdir tmp/first/second && touch tmp/first/second/example.file
  ./cache store --key restore-dir-hierarchy --path tmp/first/second
  rm -rf tmp

  run ./cache has_key restore-dir-hierarchy
  assert_success

  run ./cache restore --key restore-dir-hierarchy

  assert_success
  assert [ -e "tmp/first/second/example.file" ]
  assert_output --partial "Using cache key: restore-dir-hierarchy."
  assert_output --partial "Transferring from cache store, using cache key: restore-dir-hierarchy."
  assert_output --partial "Transfer completed."
}

@test "restoring nonexistent directory from cache" {
  run ./cache has_key test
  assert_failure

  run ./cache restore --key test

  assert_success
  assert_output --partial "Using cache key: test".
  assert_output --partial "Key 'test' does not exist in the cache store."
}

################################################################################
# cache clear
################################################################################

@test "emptying cache store when it isn't empty" {
  mkdir tmp && touch tmp/example.file
  ./cache store --key test-emptying --path tmp

  run ./cache is_not_empty
  assert_success

  run ./cache clear

  assert_success
  assert_output --partial "Cache is empty."
  refute_output --partial "Usage: rm [-r] [-f] files..."

}

@test "emptying cache store when cache is empty" {
  run ./cache is_not_empty
  assert_failure

  run ./cache clear

  assert_success
  assert_output --partial "Cache is empty."
  refute_output --partial "Usage: rm [-r] [-f] files..."
}

################################################################################
# cache list
################################################################################

@test "listing cache store when it has cached keys" {
  mkdir tmp && touch tmp/example.file
  ./cache store --key listing-v1 --path tmp
  ./cache store --key listing-v2 --path tmp

  run ./cache is_not_empty
  assert_success

  run ./cache has_key listing-v1
  assert_success

  run ./cache has_key listing-v2
  assert_success

  run ./cache list

  assert_success
  assert_output --partial "listing-v1"
  assert_output --partial "listing-v2"
}

@test "listing cache keys when cache is empty" {
  ./cache clear

  run ./cache is_not_empty
  assert_failure

  run ./cache list
  assert_success
}

################################################################################
# cache has_key
################################################################################

@test "checking if an existing key is present in cache store" {
  mkdir tmp && touch tmp/example.file
  ./cache store --key example-key --path tmp

  run ./cache is_not_empty
  assert_success

  run ./cache has_key example-key

  assert_success
  assert_output --partial "Key example-key exists in the cache store."
}

@test "checking if nonexistent key is present in empty cache store" {
  run ./cache clear
  assert_success

  run ./cache is_not_empty
  assert_failure

  run ./cache has_key example-key

  assert_failure
  assert_output --partial "Checking if key example-key is present in cache store."
  assert_output --partial "Key example-key doesn't exist in the cache store."
}

################################################################################
# cache is_not_empty
################################################################################

@test "is_not_empty should fail when cache store is empty" {
  ./cache clear

  run ./cache is_not_empty
  assert_failure
}

@test "is_not_empty should not fail  when cache is not empty" {
  ./cache store --key semaphore --path .semaphore

  run ./cache list
  assert_output --partial "semaphore"

  run ./cache is_not_empty
  assert_success
}
