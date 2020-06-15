#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  rm -rf semaphore-demo-*
}


################################################################################
# cache autostore/autorestore
################################################################################

@test "[macOS] cache - autostore/autorestore [bundle]" {

  git clone https://github.com/Shopify/example-ruby-app
  cd example-ruby-app
  rm -rf .ruby-version && rm -rf Gemfile.lock
  bundle install --path vendor/bundle > /dev/null

  run cache store

  assert_success
  assert_output --partial "* Detected Gemfile.lock."
  assert_output --partial "Upload complete."

  rm -rf vendor/bundle

  run cache restore
  assert_success
  assert_output --partial  "* Fetching 'vendor/bundle' directory with cache keys"
  assert_output --partial "Restored: vendor/bundle/"

  run cache delete gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock)
  cd ../
  rm -rf example-ruby-app
}

@test "[macOS] cache - autostore/autorestore [nodejs]" {

  git clone git@github.com:semaphoreci-demos/semaphore-demo-javascript.git
  cd semaphore-demo-javascript/src/client/

  npm install > /dev/null

  run cache store

  assert_success
  assert_output --partial "* Detected package-lock.json"
  assert_output --partial "Upload complete."

  rm -rf node_modules

  run cache restore
  assert_success
  assert_output --partial "* Fetching 'node_modules' directory with cache keys"
  assert_output --partial "Restored: node_modules/"

  run cache delete node-modules-$SEMAPHORE_GIT_BRANCH-$(checksum package-lock.json)
  cd ../../../
  rm -rf semaphore-demo-javascript
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

  run cache delete pods-$SEMAPHORE_GIT_BRANCH-$(checksum Podfile.lock)
  cd ../../../
  rm -rf example-app-ios.git
}
