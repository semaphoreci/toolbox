#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  echo "hello" > /tmp/unique-file-$SEMAPHORE_JOB_ID
}

@test "artifacts - uploading to project level" {
  run artifact push project /tmp/unique-file-$SEMAPHORE_JOB_ID
  assert_success

  run artifact pull project unique-file-$SEMAPHORE_JOB_ID
  assert_success
  run cat unique-file-$SEMAPHORE_JOB_ID
  assert_output "hello"

  run artifact yank project unique-file-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to project level using stdin" {
  run echo "from stdin" | artifact push project - -d from-stdin-$SEMAPHORE_JOB_ID
  assert_success

  run artifact pull project from-stdin-$SEMAPHORE_JOB_ID
  assert_success
  run cat from-stdin-$SEMAPHORE_JOB_ID
  assert_output "from stdin"

  run artifact yank project from-stdin-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to workflows level" {
  run artifact push workflows /tmp/unique-file-$SEMAPHORE_JOB_ID
  assert_success

  run artifact pull project unique-file-$SEMAPHORE_JOB_ID
  assert_success
  run cat unique-file-$SEMAPHORE_JOB_ID
  assert_output "hello"

  run artifact yank workflows unique-file-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to workflows level using stdin" {
  run echo "from stdin" | artifact push workflows - -d from-stdin-$SEMAPHORE_JOB_ID
  assert_success

  run artifact pull workflows from-stdin-$SEMAPHORE_JOB_ID
  assert_success
  run cat from-stdin-$SEMAPHORE_JOB_ID
  assert_output "from stdin"

  run artifact yank workflows from-stdin-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to job level" {
  run artifact push job /tmp/unique-file-$SEMAPHORE_JOB_ID
  assert_success

  run artifact pull project unique-file-$SEMAPHORE_JOB_ID
  assert_success
  run cat unique-file-$SEMAPHORE_JOB_ID
  assert_output "hello"

  run artifact yank job unique-file-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to job level using stdin" {
  run echo "from stdin" | artifact push job - -d from-stdin-$SEMAPHORE_JOB_ID
  assert_success

  run artifact pull job from-stdin-$SEMAPHORE_JOB_ID
  assert_success
  run cat from-stdin-$SEMAPHORE_JOB_ID
  assert_output "from stdin"

  run artifact yank job from-stdin-$SEMAPHORE_JOB_ID
  assert_success
}
