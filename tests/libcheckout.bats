#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_DIR="repo"
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

@test "libcheckout - Checkout Tag" {
  export SEMAPHORE_GIT_BRANCH='v2.5.0'
  export SEMAPHORE_GIT_SHA=7219ef6

  run checkout
  assert_success
  assert_output --partial "Performing shallow clone with depth: 50"
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
  refute_output --partial "SHA: $SEMAPHORE_GIT_SHA not found performing full clone: command not found"

}

@test "libcheckout - Checkout refs/tags" {
  export SEMAPHORE_GIT_BRANCH='refs/tags/v2.5.0'
  export SEMAPHORE_GIT_SHA=7219ef6

  run checkout
  assert_success
  assert_output --partial "Performing shallow clone with depth: 50"
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"

}

@test "libcheckout - Checkout nonexisting SHA" {
  export SEMAPHORE_GIT_SHA=1234567

  run checkout
  assert_failure
}

@test "libcheckout - Checkout use cache" {

  cache clear

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


}


@test "libcheckout - Checkout and use cache" {

  export SEMAPHORE_GIT_URL="https://github.com/rails/rails.git"
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_DIR=rails
  export SEMAPHORE_GIT_SHA=f907b418aecfb6dab4e30149b88a8593ddd321b9
  cache clear

  run checkout
  assert_success

  export SEMAPHORE_GIT_BRANCH=5-0-stable
  cd ~
  cache list
  rm -rf $SEMAPHORE_GIT_DIR

  run checkout --use-cache
  assert_success
  assert_output --partial "MISS: git-cache-"
  assert_output --partial "No git cache... caching"
  refute_output --partial "HIT: git-cache-"

  cache clear
}
