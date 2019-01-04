#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_DIR=/tmp/repo
  export SEMAPHORE_GIT_SHA=5608567

  source ~/.toolbox/libcheckout
  rm -rf $SEMAPHORE_GIT_DIR
}

teardown() {
  rm -rf $SEMAPHORE_GIT_DIR
}

@test "libcheckout - Checkout repository" {
  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
}

@test "libcheckout - Checkout old revision" {
  export SEMAPHORE_GIT_BRANCH=patch-id
  export SEMAPHORE_GIT_SHA=da70719

  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"

}

@test "libcheckout - Checkout nonexisting SHA" {
  export SEMAPHORE_GIT_SHA=1234567

  run checkout
  assert_failure
}

@test "libcheckout - Checkout use cache" {
  ./cache delete $(cache list 2>&1 | grep git-cache- | awk '{ print $1 }')

  run checkout --use-cache
  assert_success
  assert_output --partial "MISS: git-cache-"
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
  assert_output --partial "No git cache... caching"
  refute_output --partial "HIT: git-cache-"

}

@test "libcheckout - Checkout restore from cache" {

  run checkout --use-cache
  assert_success
  assert_output --partial "HIT: git-cache-"
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
  refute_output --partial "MISS: git-cache-"

}

@test "libcheckout - Checkout cache outdated" {
  export SEMAPHORE_GIT_CACHE_AGE=1

  run checkout --use-cache
  assert_success
  assert_output --partial "HIT: git-cache-"
  assert_output --partial "Git cache outdated, refreshing..."
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
  refute_output --partial "MISS: git-cache-"

  ./cache delete $(cache list 2>&1 | grep git-cache- | awk '{ print $1 }')

}
