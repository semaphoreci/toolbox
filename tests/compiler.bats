#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

@test "compiler can evaluare change_in expressions" {
  rm -rf /tmp/test-repo
  cp -R tests/compiler/test-repo /tmp/test-repo

  cd /tmp/test-repo
  git config --global user.email "you@example.com"
  git config --global user.name "Your Name"
  git init .
  git add .
  git commit -m "Bootstrap"

  run spc --input .semaphore/semaphore.yml --output .semaphore/semaphore.yml.compiler --logs .semaphore/semaphore.yml.logs
  assert_success

  run cat .semaphore/semaphore.yml.compiler
  assert_output "adasdas"
}
