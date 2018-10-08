#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  ./cache clear
  rm -rf tmp
}

@test "fallback key prototype" {
  ./cache store --key v1-gems-master-p12q13r34 --path tests

  run ./cache restore --key v1-gems-master-2new99666,v1-gems-master-*

  assert_success
  assert_output --partial "Using cache key: v1-gems-master-2new99666".
  assert_output --partial "Key 'v1-gems-master-p12q13r34' does not exist in the cache store."
  assert_output --partial "Fallbacking to 'v1-gems-master-*'"
  assert_output --partial "Most recent fallback found 'v1-gems-master-p12q13r34'"
  refute_output --partial "/home/semaphore/toolbox"
}

@test "fallbacsk key prototype" {
  ./cache store --key v1-gems-master-p12q13r34 --path tests

  run ./cache restore --key v1-gems-master-2new99666,v1-gems-master-*

  assert_success
  assert_output --partial "Using cache key: v1-gems-master-2new99666".
  assert_output --partial "Key 'v1-gems-master-p12q13r34' does not exist in the cache store."
  assert_output --partial "Fallbacking to 'v1-gems-master-*'"
  assert_output --partial "Most recent fallback found 'v1-gems-master-p12q13r34'"
  refute_output --partial "/home/semaphore/toolbox"
}
