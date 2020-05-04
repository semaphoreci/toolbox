#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  unset SEMAPHORE_GIT_REF_TYPE
  unset SEMAPHORE_GIT_TAG_NAME
  unset SEMAPHORE_GIT_PR_SLUG
  unset SEMAPHORE_GIT_PR_NAME
  unset SEMAPHORE_GIT_PR_NUMBER

  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_DIR="repo"
  export SEMAPHORE_GIT_SHA=5608567
  export SEMAPHORE_GIT_REPO_SLUG="mojombo/grit"
  export SEMAPHORE_GIT_REF="refs/heads/master"

  source ~/.toolbox/libcheckout
  rm -rf $SEMAPHORE_GIT_DIR
}

teardown() {
  rm -rf $SEMAPHORE_GIT_DIR
}

# Push

@test "libcheckout - [Push]" {
  export SEMAPHORE_GIT_REF_TYPE="push"
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_success
  assert_output --partial "HEAD is now at 91940c2"
}

@test "libcheckout - [Push] missing sha" {
  export SEMAPHORE_GIT_REF_TYPE="push"
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a4

  run checkout
  assert_failure
}

@test "libcheckout - [Push] missing branch" {
  export SEMAPHORE_GIT_REF_TYPE="push"
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_success
}

# Tag

@test "libcheckout - [Tag]" {
  export SEMAPHORE_GIT_REF_TYPE="tag"
  export SEMAPHORE_GIT_TAG_NAME='v2.4.1'
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA Release $SEMAPHORE_GIT_TAG_NAME"
}

@test "libcheckout - [Tag] missing tag" {
  export SEMAPHORE_GIT_REF_TYPE="tag"
  export SEMAPHORE_GIT_TAG_NAME='v9.4.1'
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_failure
  assert_output --partial "Release $SEMAPHORE_GIT_TAG_NAME not found .... Exiting"
}

# PR

@test "libcheckout - [PR]" {
  export SEMAPHORE_GIT_REF_TYPE="pull-request"
  export SEMAPHORE_GIT_REF="refs/pull/186/merge"
  export SEMAPHORE_GIT_SHA=30774365e11f2b1e18706c9ed0920369f6d7c205

  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
}

@test "libcheckout - [PR] no ref" {
  export SEMAPHORE_GIT_REF_TYPE="pull-request"
  export SEMAPHORE_GIT_REF="refs/pull/1111/merg"

  run checkout
  assert_failure
  assert_output --partial "Revision: $SEMAPHORE_GIT_SHA not found .... Exiting"
}

# noRefType

@test "libcheckout - [noRef] Checkout repository" {
  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
}

@test "libcheckout - [noRef] Checkout old revision" {
  export SEMAPHORE_GIT_BRANCH=patch-id
  export SEMAPHORE_GIT_SHA=da70719

  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
}

@test "libcheckout - [noRef] Checkout Tag" {
  export SEMAPHORE_GIT_BRANCH='v2.5.0'
  export SEMAPHORE_GIT_SHA=7219ef6

  run checkout
  assert_success
  assert_output --partial "Performing shallow clone with depth: 50"
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
  refute_output --partial "SHA: $SEMAPHORE_GIT_SHA not found performing full clone: command not found"
}

@test "libcheckout - [noRef] Checkout refs/tags" {
  export SEMAPHORE_GIT_BRANCH='refs/tags/v2.5.0'
  export SEMAPHORE_GIT_SHA=7219ef6

  run checkout
  assert_success
  assert_output --partial "Performing shallow clone with depth: 50"
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
}

@test "libcheckout - [noRef] Checkout nonexisting SHA" {
  export SEMAPHORE_GIT_SHA=1234567

  run checkout
  assert_failure
}

@test "libcheckout - [noRef] Checkout use cache" {

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

@test "libcheckout - Checkout and use-cache nonexisting SHA" {
  export SEMAPHORE_GIT_SHA=1234567
  export SEMAPHORE_GIT_BRANCH=master
  cd ~
  rm -rf $SEMAPHORE_GIT_DIR

  run checkout --use-cache
  assert_failure
  assert_output --partial "Revision: $SEMAPHORE_GIT_SHA not found"
}
