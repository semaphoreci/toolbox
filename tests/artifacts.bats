#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  echo "hello" > /tmp/unique-file-$SEMAPHORE_JOB_ID
}

teardown() {
  rm -rf /tmp/unique-file-$SEMAPHORE_JOB_ID
  rm -rf from-stdin-$SEMAPHORE_JOB_ID
  rm -rf unique-file-$SEMAPHORE_JOB_ID
  rm -rf image-$SEMAPHORE_JOB_ID.gz
  rm -rf image-$SEMAPHORE_JOB_ID
}

@test "artifacts - uploading to project level" {
  run artifact push project /tmp/unique-file-$SEMAPHORE_JOB_ID -v
  assert_success

  run artifact pull project unique-file-$SEMAPHORE_JOB_ID -v
  assert_success
  run cat unique-file-$SEMAPHORE_JOB_ID
  assert_output "hello"

  run artifact yank project unique-file-$SEMAPHORE_JOB_ID -v
  assert_success
}

@test "artifacts - uploading to project level using stdin" {
  run bash -c "echo 'from stdin' | artifact push project - -d from-stdin-$SEMAPHORE_JOB_ID -v"
  assert_success

  run artifact pull project from-stdin-$SEMAPHORE_JOB_ID -v
  assert_success

  run cat from-stdin-$SEMAPHORE_JOB_ID
  assert_output "from stdin"

  run artifact yank project from-stdin-$SEMAPHORE_JOB_ID -v
  assert_success
}

@test "artifacts - uploading bigger file to project level using stdin" {
  run bash -c "docker image save alpine/helm | gzip | artifact push project - -d image-$SEMAPHORE_JOB_ID.gz -v"
  assert_success

  run artifact pull project image-$SEMAPHORE_JOB_ID.gz -v
  assert_success

  run gzip -d image-$SEMAPHORE_JOB_ID.gz
  assert_success

  run artifact yank project image-$SEMAPHORE_JOB_ID.gz -v
  assert_success
}

@test "artifacts - uploading to workflow level" {
  run artifact push workflow /tmp/unique-file-$SEMAPHORE_JOB_ID -v
  assert_success

  run artifact pull workflow unique-file-$SEMAPHORE_JOB_ID -v
  assert_success
  run cat unique-file-$SEMAPHORE_JOB_ID
  assert_output "hello"

  run artifact yank workflow unique-file-$SEMAPHORE_JOB_ID
  assert_success
}

@test "artifacts - uploading to workflows level using stdin" {
  run bash -c "echo 'from stdin' | artifact push workflow - -d from-stdin-$SEMAPHORE_JOB_ID -v"
  assert_success

  run artifact pull workflow from-stdin-$SEMAPHORE_JOB_ID -v
  assert_success
  run cat from-stdin-$SEMAPHORE_JOB_ID
  assert_output "from stdin"

  run artifact yank workflow from-stdin-$SEMAPHORE_JOB_ID -v
  assert_success
}

@test "artifacts - uploading bigger file to workflow level using stdin" {
  run bash -c "docker image save alpine/helm | gzip | artifact push workflow - -d image-$SEMAPHORE_JOB_ID.gz -v"
  assert_success

  run artifact pull workflow image-$SEMAPHORE_JOB_ID.gz -v
  assert_success

  run gzip -d image-$SEMAPHORE_JOB_ID.gz
  assert_success

  run artifact yank workflow image-$SEMAPHORE_JOB_ID.gz -v
  assert_success
}

@test "artifacts - uploading to job level" {
  run artifact push job /tmp/unique-file-$SEMAPHORE_JOB_ID -v
  assert_success

  run artifact pull job unique-file-$SEMAPHORE_JOB_ID -v
  assert_success
  run cat unique-file-$SEMAPHORE_JOB_ID
  assert_output "hello"

  run artifact yank job unique-file-$SEMAPHORE_JOB_ID -v
  assert_success
}

@test "artifacts - uploading to job level using stdin" {
  run bash -c "echo 'from stdin' | artifact push job - -d from-stdin-$SEMAPHORE_JOB_ID -v"
  assert_success

  run artifact pull job from-stdin-$SEMAPHORE_JOB_ID -v
  assert_success
  run cat from-stdin-$SEMAPHORE_JOB_ID
  assert_output "from stdin"

  run artifact yank job from-stdin-$SEMAPHORE_JOB_ID -v
  assert_success
}

@test "artifacts - uploading bigger file to job level using stdin" {
  run bash -c "docker image save alpine/helm | gzip | artifact push job - -d image-$SEMAPHORE_JOB_ID.gz -v"
  assert_success

  run artifact pull job image-$SEMAPHORE_JOB_ID.gz -v
  assert_success

  run gzip -d image-$SEMAPHORE_JOB_ID.gz
  assert_success

  run artifact yank job image-$SEMAPHORE_JOB_ID.gz -v
  assert_success
}