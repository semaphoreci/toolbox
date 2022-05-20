#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  echo "hello" > /tmp/unique-file-$SEMAPHORE_JOB_ID
}

@test "artifacts - uploading to proect level" {
  run artifact push project /tmp/unique-file-$SEMAPHORE_JOB_ID
  assert_success


  run artifact yank project unique-file-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to workflows level" {
  run artifact push workflows /tmp/unique-file-$SEMAPHORE_JOB_ID
  assert_success

  run artifact yank workflows unique-file-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to job level" {
  run artifact push job /tmp/unique-file-$SEMAPHORE_JOB_ID
  assert_success

  run artifact yank job unique-file-$SEMAPHORE_JOB_ID
  assert_success
}
