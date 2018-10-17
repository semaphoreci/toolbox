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
  run ./cache store test-storing tmp

  assert_success
  assert_line "Uploading 'tmp' with cache key 'test-storing'..."
  assert_line "Upload complete."
  refute_line "test-storing"

  run ./cache has_key test-storing
  assert_success
}

@test "save local file to cache store with normalized key" {
  mkdir tmp && touch tmp/example.file
  run ./cache store test/storing tmp

  assert_success
  assert_line "Key test/storing is normalized to test-storing."
  assert_line "Uploading 'tmp' with cache key 'test-storing'..."
  assert_line "Upload complete."
  refute_line "test-storing"

  run ./cache has_key test-storing
  assert_success
}

@test "save nonexistent local file to cache store" {
  run ./cache store test-storing tmp

  assert_success
  assert_line "'tmp' doesn't exist locally."
}

@test "store with key which is already present in cache" {
  mkdir tmp && touch tmp/example.file
  ./cache store test-storing tmp
  run ./cache has_key test-storing
  assert_success

  run ./cache store test-storing tmp

  assert_success
  assert_line "Key 'test-storing' already exists."
  refute_line "test-storing"

  run ./cache has_key test-storing
  assert_success
}

################################################################################
# cache restore
################################################################################

@test "restoring existing directory from cache and perserving the directory hierarchy" {
  mkdir tmp && mkdir tmp/first && mkdir tmp/first/second && touch tmp/first/second/example.file
  ./cache store restore-dir-hierarchy tmp/first/second
  rm -rf tmp

  run ./cache has_key restore-dir-hierarchy
  assert_success

  run ./cache restore restore-dir-hierarchy

  assert_success
  assert [ -e "tmp/first/second/example.file" ]
  assert_line "HIT: restore-dir-hierarchy, using key restore-dir-hierarchy"
  assert_output --partial "Restored: tmp/first/second/"
  refute_output --partial "/home/semaphore/toolbox"
}

@test "restors the key if it is available" {
  touch tmp.file
  ./cache store tmp1 tmp.file
  ./cache store tmp12 tmp.file

  run ./cache restore tmp1

  assert_success
  assert_line "HIT: tmp1, using key tmp1"
  refute_output --partial "HIT: tmp1, using key tmp12"
  assert_output --partial "Restored: tmp.file"
  refute_output --partial "/home/semaphore/toolbox"
}

@test "restoring nonexistent directory from cache" {
  run ./cache has_key test
  assert_failure

  run ./cache restore test

  assert_success
  assert_line "MISS: test"
  refute_output --partial "/home/semaphore/toolbox"
}

@test "fallback key prototype" {
  touch tmp.file
  ./cache store v1-gems-master-p12q13r34 tmp.file

  run ./cache restore v1-gems-master-2new99666,v1-gems-master-*

  assert_success
  assert_line "MISS: v1-gems-master-2new99666"
  assert_line "HIT: v1-gems-master-*, using key v1-gems-master-p12q13r34"
  assert_line "Restored: tmp.file"
  refute_output --partial "/home/semaphore/toolbox"
}

@test "fallback key prototype uses normalized keys" {
  touch tmp.file
  ./cache store modules-ms/quick-update tmp.file

  run ./cache restore modules-master-1234,modules-ms/quick-update

  assert_success
  assert_line "Key modules-ms/quick-update is normalized to modules-ms-quick-update."
  assert_line "HIT: modules-ms-quick-update, using key modules-ms-quick-update"
  assert_output --partial "Restored: tmp.file"
  refute_output --partial "/home/semaphore/toolbox"
}

################################################################################
# cache clear
################################################################################

@test "emptying cache store when it isn't empty" {
  mkdir tmp && touch tmp/example.file
  ./cache store test-emptying tmp

  run ./cache is_not_empty
  assert_success

  run ./cache clear

  assert_success
  assert_output --partial "Deleted all caches."
  refute_output --partial "Usage: rm [-r] [-f] files"

}

@test "emptying cache store when cache is empty" {
  run ./cache is_not_empty
  assert_failure

  run ./cache clear

  assert_success
  assert_line "Deleted all caches."
  refute_output --partial "Usage: rm [-r] [-f] files"
}

################################################################################
# cache list
################################################################################

@test "listing cache store when it has cached keys" {
  mkdir tmp && touch tmp/example.file
  ./cache store ms/quick-update tmp
  ./cache store listing-v2 tmp

  run ./cache is_not_empty
  assert_success

  run ./cache has_key ms/quick-update
  assert_success

  run ./cache has_key listing-v2
  assert_success

  run ./cache list

  assert_success
  assert_output --partial "ms-quick-update"
  assert_output --partial "listing-v2"
}

@test "listing cache keys when cache is empty" {
  ./cache clear

  run ./cache is_not_empty
  assert_failure

  run ./cache list
  assert_success
  assert_line "Cache is empty."
}

################################################################################
# cache has_key
################################################################################

@test "checking if an existing key is present in cache store" {
  mkdir tmp && touch tmp/example.file
  ./cache store example-key tmp

  run ./cache is_not_empty
  assert_success

  run ./cache has_key example-key

  assert_success
  assert_output --partial "Key example-key exists in the cache store."
}

@test "checking if an existing key with / is present in cache store" {
  mkdir tmp && touch tmp/example.file
  ./cache store ek/quick-update tmp

  run ./cache is_not_empty
  assert_success

  run ./cache has_key ek/quick-update

  assert_success
  assert_line "Key ek/quick-update is normalized to ek-quick-update."
  assert_output --partial "Key ek-quick-update exists in the cache store."

  run ./cache has_key ek-quick-update

  assert_success
  assert_output --partial "Key ek-quick-update exists in the cache store."
}

@test "checking if nonexistent key is present in empty cache store" {
  run ./cache clear
  assert_success

  run ./cache is_not_empty
  assert_failure

  run ./cache has_key example-key

  assert_failure
  assert_output --partial "Key example-key doesn't exist in the cache store."
}

################################################################################
# cache delete
################################################################################

@test "deletion of an existing key" {
  mkdir tmp && touch tmp/example.file
  ./cache store example-key tmp
  ./cache has_key example-key

  run ./cache delete example-key

  assert_success
  assert_output --partial "Key example-key is deleted."

  run ./cache has_key example-key
  assert_failure
}

@test "delition of an existing key with /" {
  mkdir tmp && touch tmp/example.file
  ./cache store ek/quick-update tmp

  run ./cache is_not_empty
  assert_success

  run ./cache delete ek/quick-update

  assert_success
  assert_line "Key ek/quick-update is normalized to ek-quick-update."
  assert_output --partial "Key ek-quick-update is deleted."

  run ./cache has_key ek/quick-update

  assert_failure
  assert_line "Key ek/quick-update is normalized to ek-quick-update."
  assert_output --partial "Key ek-quick-update doesn't exist in the cache store."
}

@test "deletion of a nonexistent key" {
  run ./cache has_key example-key
  assert_failure

  run ./cache delete example-key

  assert_success
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

@test "is_not_empty should not fail when cache is not empty" {
  ./cache store semaphore .semaphore

  run ./cache list
  assert_output --partial "semaphore"

  run ./cache is_not_empty
  assert_success
}

################################################################################
# cache usage
################################################################################

@test "communicates the correct cache usage" {
  dd if=/dev/zero of=file.tmp bs=1M count=50
  ./cache store tmp file.tmp
  run ./cache usage

  assert_success
  assert_line "FREE SPACE: 9.6G"
  assert_line "USED SPACE: 50K"

  rm -f file.tmp
}
