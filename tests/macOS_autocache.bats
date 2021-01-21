#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

PROJECT_ROOT=$(pwd)

setup() {
  cache clear
  cd $PROJECT_ROOT
}

teardown() {
  git reset --hard
  git clean -fd
}

@test "[macOS] cache - autostore/autorestore [bundle]" {
  cd tests/autocache/ruby
  bundle install --path vendor/bundle

  run cache store

  assert_success
  assert_output --partial "* Detected Gemfile.lock."
  assert_output --partial "Upload complete."

  rm -rf vendor/bundle

  run cache restore
  assert_success
  assert_output --partial  "* Fetching 'vendor/bundle' directory with cache keys"
  assert_output --partial "Restored: vendor/bundle/"
}

@test "[macOS] cache - autostore/autorestore [nodejs]" {
  cd tests/autocache/js
  npm install

  run cache store

  assert_success
  assert_output --partial "* Detected package-lock.json"
  assert_output --partial "Upload complete."

  rm -rf node_modules

  run cache restore
  assert_success
  assert_output --partial "* Fetching 'node_modules' directory with cache keys"
  assert_output --partial "Restored: node_modules/"
}

@test "[macOS] cache - autostore/autorestore [Pods]" {
  git clone https://github.com/particle-iot/example-app-ios.git
  cd example-app-ios

  export LANG=en_US.UTF-8

  pod install 1>/dev/null 2>&1

  run cache store

  assert_success
  assert_output --partial "* Detected Podfile.lock"
  assert_output --partial "Upload complete."

  rm -rf Pods

  run cache restore
  assert_success
  assert_output --partial "* Fetching 'Pods' directory with cache keys"
  assert_output --partial "Restored: Pods/"
}
