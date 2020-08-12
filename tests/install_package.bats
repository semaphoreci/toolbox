#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

@test "test install-package" {
  run install-package mtr
  assert_success
}
@test "test install_package files" {
  run ls -lah ~/.deb-cache
  assert_success
  assert_output --partial "mtr"
}
@test "test caching of packages" {
  cache has_key install_package_cache
  assert_success
}