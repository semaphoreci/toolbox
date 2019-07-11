#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  rm -rf tmp
  ./cache delete bats-test-$SEMAPHORE_GIT_BRANCH
  ./cache delete bats-test-$SEMAPHORE_GIT_BRANCH-1
  unset CACHE_SIZE
}

normalize_key() {
  local word
  local result
  word=$1

  result=${word//\//-}
  echo "$result"
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
  refute_output --partial "command not found"
}

################################################################################
# cache store
################################################################################

@test "save local file to cache store" {
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && touch tmp/example.file

  run ./cache store bats-test-$SEMAPHORE_GIT_BRANCH tmp

  assert_success
  assert_line "Uploading 'tmp' with cache key '${test_key}'..."
  assert_line "Upload complete."
  refute_line ${test_key}
  refute_output --partial "command not found"

  run ./cache has_key bats-test-$SEMAPHORE_GIT_BRANCH

  assert_line "Key ${test_key} exists in the cache store."
  assert_success
  refute_output --partial "command not found"

  # retry and ssh-session-cli are changed by setup commands in the job environment
  git checkout retry ssh-session-cli
  run git status

  assert_output --partial "nothing to commit"
}

@test "save local file to cache store with normalized key" {
  test_key=$(normalize_key bats/test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && touch tmp/example.file

  run ./cache store bats/test-$SEMAPHORE_GIT_BRANCH tmp

  assert_success
  assert_line "Key bats/test-${SEMAPHORE_GIT_BRANCH} is normalized to ${test_key}."
  assert_line "Uploading 'tmp' with cache key '${test_key}'..."
  assert_line "Upload complete."
  refute_line ${test_key}
  refute_output --partial "command not found"

  run ./cache has_key bats/test-$SEMAPHORE_GIT_BRANCH

  assert_success
  refute_output --partial "command not found"
}

@test "save nonexistent local file to cache store" {
  run ./cache store test-storing tmp

  assert_success
  assert_line "'tmp' doesn't exist locally."
  refute_output --partial "Cannot stat: No such file or directory"
  refute_output --partial "command not found"
}

@test "store with key which is already present in cache" {
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && touch tmp/example.file
  ./cache store bats-test-$SEMAPHORE_GIT_BRANCH tmp

  run ./cache has_key bats-test-$SEMAPHORE_GIT_BRANCH
  assert_success
  refute_output --partial "command not found"

  run ./cache store bats-test-$SEMAPHORE_GIT_BRANCH tmp

  assert_success
  assert_line "Key '${test_key}' already exists."
  refute_line "Uploading 'tmp' with cache key '${tes_key}'..."
  refute_line ${test_key}
  refute_output --partial "command not found"

  run ./cache has_key bats-test-$SEMAPHORE_GIT_BRANCH
  assert_success
  refute_output --partial "command not found"
}

@test "automatic key deletion in case of insufficient space" {
  preexisting_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  new_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH-1)
  dd if=/dev/zero of=tmp.file bs=1M count=50
  dd if=/dev/zero of=tmp.larger_file bs=1M count=70
  export CACHE_SIZE=110
  ./cache store $preexisting_key tmp.file
  ./cache store tmp-key tmp.file
  ./cache list

  run ./cache store $new_key tmp.larger_file
  assert_line "Not enough space, deleting the oldest keys."
  assert_line "Key ${preexisting_key} is deleted."
  assert_line "Key tmp-key is deleted."
  assert_line "Uploading 'tmp.larger_file' with cache key '${new_key}'..."
  refute_output --partial "command not found"

  run ./cache has_key tmp-key
  assert_failure
  refute_output --partial "command not found"

  run ./cache has_key $preexisting_key
  assert_failure
  refute_output --partial "command not found"

  run ./cache has_key $new_key
  assert_success
  refute_output --partial "command not found"

  ./cache delete tmp-key
}

@test "storing key that exceeds the allowed disk space size" {
  raw_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH-1)
  dd if=/dev/zero of=tmp.file bs=1M count=70
  export CACHE_SIZE=50

  run ./cache store $test_key tmp.file
  assert_success
  assert_line "Archive exceeds allocated 50K for cache."
  refute_output --partial "command not found"
}

################################################################################
# cache restore
################################################################################

@test "restoring existing directory from cache and perserving the directory hierarchy" {
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && mkdir tmp/first && mkdir tmp/first/second && touch tmp/first/second/example.file
  ./cache store bats-test-$SEMAPHORE_GIT_BRANCH tmp/first/second
  rm -rf tmp

  run ./cache has_key bats-test-$SEMAPHORE_GIT_BRANCH
  assert_success
  refute_output --partial "command not found"

  run ./cache restore bats-test-$SEMAPHORE_GIT_BRANCH

  assert_success
  assert [ -e "tmp/first/second/example.file" ]
  assert_line "HIT: ${test_key}, using key ${test_key}"
  assert_output --partial "Restored: tmp/first/second/"
  refute_output --partial "/home/semaphore/toolbox"
  refute_output --partial "command not found"

  # retry and ssh-session-cli are changed by setup commands in the job environment
  git checkout retry ssh-session-cli
  run git status

  assert_output --partial "nothing to commit"
}

@test "restores the key if it is available" {
  test_key_1=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  test_key_2=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH-1)
  touch tmp.file
  ./cache store $test_key_1 tmp.file
  ./cache store $test_key_2 tmp.file
  refute_output --partial "command not found"

  run ./cache restore $test_key_1

  assert_success
  assert_line "HIT: ${test_key_1}, using key ${test_key_1}"
  refute_output --partial "HIT: ${test_key_1}, using key ${test_key_2}"
  assert_output --partial "Restored: tmp.file"
  refute_output --partial "/home/semaphore/toolbox"
  refute_output --partial "command not found"
}

@test "restoring nonexistent directory from cache" {
  run ./cache has_key test-12123
  assert_failure

  run ./cache restore test-12123

  assert_success
  assert_line "MISS: test-12123"
  refute_output --partial "/home/semaphore/toolbox"
  refute_output --partial "command not found"
}

@test "fallback key option" {
  touch tmp.file
  test_key_1=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  test_key_2=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH-1)
  ./cache store bats-test-$SEMAPHORE_GIT_BRANCH tmp.file

  run ./cache restore bats-test-$SEMAPHORE_GIT_BRANCH-1,bats-test

  assert_success
  assert_line "MISS: ${test_key_2}"
  assert_line "HIT: bats-test, using key ${test_key_1}"
  assert_line "Restored: tmp.file"
  refute_output --partial "/home/semaphore/toolbox"
  refute_output --partial "command not found"
}

@test "fallback key option uses normalized keys" {
  touch tmp.file
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  ./cache store bats/test-$SEMAPHORE_GIT_BRANCH tmp.file

  run ./cache restore modules-master-1234,bats/test-$SEMAPHORE_GIT_BRANCH

  assert_success
  assert_line "Key bats/test-$SEMAPHORE_GIT_BRANCH is normalized to ${test_key}."
  assert_line "HIT: ${test_key}, using key ${test_key}"
  assert_output --partial "Restored: tmp.file"
  refute_output --partial "/home/semaphore/toolbox"
  refute_output --partial "command not found"
}

################################################################################
# cache clear
################################################################################

@test "emptying cache store when it isn't empty" {
  if [ "$SEMAPHORE_GIT_BRANCH" != "master" ]; then
    skip "- - avoiding cache clear on non master branch"
  fi

  mkdir tmp && touch tmp/example.file
  ./cache store test-emptying tmp

  run ./cache is_not_empty
  assert_success
  refute_output --partial "command not found"

  run ./cache clear

  assert_success
  assert_output --partial "Deleted all caches."
  refute_output --partial "Usage: rm [-r] [-f] files"
  refute_output --partial "command not found"
}

@test "emptying cache store when cache is empty" {
  if [ "$SEMAPHORE_GIT_BRANCH" != "master" ]; then
    skip "- - avoiding cache clear on non master branch"
  fi

  run ./cache is_not_empty
  assert_failure
  refute_output --partial "command not found"

  run ./cache clear

  assert_success
  assert_line "Deleted all caches."
  refute_output --partial "Usage: rm [-r] [-f] files"
  refute_output --partial "command not found"
}

################################################################################
# cache list
################################################################################

@test "listing cache store when it has cached keys" {
  test_key_1=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  test_key_2=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH-1)
  mkdir tmp && touch tmp/example.file
  ./cache store $test_key_1 tmp
  ./cache store ${test_key_2} tmp

  run ./cache is_not_empty
  assert_success
  refute_output --partial "command not found"

  run ./cache has_key ${test_key_1}
  assert_success
  refute_output --partial "command not found"

  run ./cache has_key ${test_key_2}
  assert_success
  refute_output --partial "command not found"

  run ./cache list

  assert_success
  assert_output --partial $test_key_1
  assert_output --partial $test_key_2
  refute_output --partial "command not found"
}

@test "listing cache keys when cache is empty" {
  if [ "$SEMAPHORE_GIT_BRANCH" != "master" ]; then
    skip "- avoiding cache clear on non master branch"
  fi

  ./cache clear

  run ./cache is_not_empty
  assert_failure
  refute_output --partial "command not found"

  run ./cache list
  assert_success
  assert_line "Cache is empty."
  refute_output --partial "command not found"
}

################################################################################
# cache has_key
################################################################################

@test "checking if an existing key is present in cache store" {
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && touch tmp/example.file
  ./cache store $test_key tmp

  run ./cache is_not_empty
  assert_success
  refute_output --partial "command not found"

  run ./cache has_key $test_key

  assert_success
  assert_output --partial "Key ${test_key} exists in the cache store."
  refute_output --partial "command not found"
}

@test "checking if an existing key with / is present in cache store" {
  test_key=$(normalize_key bats/test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && touch tmp/example.file
  ./cache store bats/test-$SEMAPHORE_GIT_BRANCH tmp

  run ./cache is_not_empty
  assert_success
  refute_output --partial "command not found"

  run ./cache has_key bats/test-$SEMAPHORE_GIT_BRANCH

  assert_success
  assert_line "Key bats/test-${SEMAPHORE_GIT_BRANCH} is normalized to ${test_key}."
  assert_output --partial "Key ${test_key} exists in the cache store."
  refute_output --partial "command not found"

  run ./cache has_key $test_key

  assert_success
  assert_output --partial "Key ${test_key} exists in the cache store."
  refute_output --partial "command not found"
}

@test "checking if nonexistent key is present in empty cache store" {
  if [ "$SEMAPHORE_GIT_BRANCH" != "master" ]; then
    skip "- avoiding cache clear on non master branch"
  fi

  run ./cache clear
  assert_success
  refute_output --partial "command not found"

  run ./cache is_not_empty
  assert_failure
  refute_output --partial "command not found"

  run ./cache has_key example-key

  assert_failure
  assert_output --partial "Key example-key doesn't exist in the cache store."
  refute_output --partial "command not found"
}

################################################################################
# cache delete
################################################################################

@test "deletion of an existing key" {
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && touch tmp/example.file
  ./cache store $test_key tmp
  ./cache has_key $test_key

  run ./cache delete $test_key

  assert_success
  assert_output --partial "Key ${test_key} is deleted."
  refute_output --partial "command not found"

  run ./cache has_key $test_key
  assert_failure
  refute_output --partial "command not found"
}

@test "deletion of an existing key with /" {
  test_key=$(normalize_key bats/test-$SEMAPHORE_GIT_BRANCH)
  mkdir tmp && touch tmp/example.file
  ./cache store bats/test-$SEMAPHORE_GIT_BRANCH tmp

  run ./cache is_not_empty
  assert_success
  refute_output --partial "command not found"

  run ./cache delete bats/test-$SEMAPHORE_GIT_BRANCH

  assert_success
  assert_line "Key bats/test-${SEMAPHORE_GIT_BRANCH} is normalized to ${test_key}."
  assert_output --partial "Key ${test_key} is deleted."
  refute_output --partial "command not found"

  run ./cache has_key bats/test-$SEMAPHORE_GIT_BRANCH

  assert_failure
  assert_line "Key bats/test-${SEMAPHORE_GIT_BRANCH} is normalized to ${test_key}."
  assert_output --partial "Key ${test_key} doesn't exist in the cache store."
  refute_output --partial "command not found"
}

@test "deletion of a nonexistent key" {
  run ./cache has_key example-nonexistent-key
  assert_failure
  refute_output --partial "command not found"

  run ./cache delete example-nonexistent-key

  assert_success
  assert_output --partial "Key example-nonexistent-key doesn't exist in the cache store."
  refute_output --partial "command not found"
}

################################################################################
# cache is_not_empty
################################################################################

@test "is_not_empty should fail when cache store is empty" {
  if [ "$SEMAPHORE_GIT_BRANCH" != "master" ]; then
    skip "- avoiding cache clear on non master branch"
  fi

  ./cache clear

  run ./cache is_not_empty
  assert_failure
  refute_output --partial "command not found"
}

@test "is_not_empty should not fail when cache is not empty" {
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  ./cache store $test_key .semaphore

  run ./cache list
  assert_output --partial "$test_key"
  refute_output --partial "command not found"

  run ./cache is_not_empty
  assert_success
  refute_output --partial "command not found"
}

################################################################################
# cache usage
################################################################################

@test "usage for empty cache store" {
  if [ "$SEMAPHORE_GIT_BRANCH" != "master" ]; then
    skip "- avoiding cache clear on non master branch"
  fi

  ./cache clear
  run ./cache usage

  assert_success
  assert_line "FREE SPACE: 9.0G"
  assert_line "USED SPACE: 0"
  refute_output --partial "command not found"

  rm -f tmp.file
}

@test "communicates the correct cache usage" {
  test_key=$(normalize_key bats-test-$SEMAPHORE_GIT_BRANCH)
  export CACHE_SIZE=100
  run ./cache usage

  dd if=/dev/zero of=tmp.file bs=1M count=50
  ./cache store $test_key tmp.file
  export CACHE_SIZE=100
  run ./cache usage

  assert_success
  assert_line "FREE SPACE: 51K"
  assert_line "USED SPACE: 50K"
  refute_output --partial "command not found"

  rm -f tmp.file
}
