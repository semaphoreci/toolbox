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

  cd $SEMAPHORE_GIT_DIR
  run bash -c "git branch | grep ${SEMAPHORE_GIT_BRANCH}"
  assert_success

  run bash -c "git rev-parse HEAD | grep ${SEMAPHORE_GIT_SHA}"
  assert_success
}

@test "libcheckout - Checkout old revision" {
  export SEMAPHORE_GIT_BRANCH=patch-id
  export SEMAPHORE_GIT_SHA=da70719

  run checkout
  assert_success

  cd $SEMAPHORE_GIT_DIR
  run bash -c "git branch | grep ${SEMAPHORE_GIT_BRANCH}"
  assert_success

  run bash -c "git rev-parse HEAD | grep ${SEMAPHORE_GIT_SHA}"
  assert_success
}

@test "libcheckout - Checkout nonexisting SHA" {
  export SEMAPHORE_GIT_SHA=1234567

  run checkout
  assert_failure
}

@test "libcheckout - Checkout use cache" {

  run checkout --use-cache
  assert_success
  assert_output --partial "MISS: git-cache-"
  assert_output --partial "No git cache... caching"

}
