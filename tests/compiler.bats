#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  unset SEMAPHORE_GIT_REF_TYPE
  unset SEMAPHORE_GIT_BRANCH
  unset SEMAPHORE_GIT_COMMIT_RANGE
  unset SEMAPHORE_GIT_SHA
  unset SEMAPHORE_MERGE_BASE
  unset SEMAPHORE_MERGE_BASE

  git config --global user.email "you@example.com"
  git config --global user.name "Your Name"

  rm -rf /tmp/test-repo-origin
  rm -rf /tmp/test-repo-clone
  cp -R tests/compiler/test-repo /tmp/test-repo-origin

  cd /tmp/test-repo-origin
  git init
  git add .
  git commit -m "Bootstrap"
  git clone /tmp/test-repo-origin /tmp/test-repo-clone
  cd -
}

@test "compiler can evaluare change_in expressions" {
  cd /tmp/test-repo-clone

  run spc evaluate change-in --input .semaphore/semaphore.yml --output .semaphore/semaphore.yml.compiler --logs .semaphore/semaphore.yml.logs
  assert_success

  run cat .semaphore/semaphore.yml.compiler
  assert_output --partial "(branch = 'master') and false"

  cd -
}
