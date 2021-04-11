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

@test "cache - autostore/autorestore with a --namespace" {
  #
  # Cache store and restore can detect and automatically create the appropriate
  # keys for any project type. For example, for ruby projects the key is:
  #
  #   gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock)
  #
  # In monorepo context, it can happen that we have we have two independent
  # services, both written in Ruby.
  #
  # We would be tempted to set up the following pipeline for this scenario:
  #
  #   blocks:
  #     - name: User Service
  #       task:
  #         jobs:
  #           - name: Tests
  #             commands:
  #               - checkout
  #               - cd services/user-service
  #               - cache restore
  #               - bundle install
  #               - cache store
  #               - bunde exec rspec rspec
  #
  #     - name: Billing Service
  #       task:
  #         jobs:
  #           - name: Tests
  #             commands:
  #               - checkout
  #               - cd services/billing-service
  #               - cache restore
  #               - bundle install
  #               - cache store
  #               - bunde exec rspec rspec
  #
  # However, if we do this, the "cache store" commands in Billing Service and
  # User Service will use the same cache key for two independent sets of gems.
  # This will lead to undesired behaviour.
  #
  # One way to solve this would be to fall back to using the long form of the
  # caching command and manually edit the keys:
  #
  # For user service:
  #
  #   cache store user-service-gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock) vendor/bundle
  #   cache restore user-service-gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock)
  #
  # For billing service:
  #
  #   cache store billing-service-gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock) vendor/bundle
  #   cache restore billing-service-gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock)
  #
  # This is a drop in overall usability and readability.
  #
  # To solve this, we are introducing a --namespace flag for the cache command.
  #
  # Example:
  #
  #   cache store --namespace user-service
  #   cache restore --namespace user-service
  #
  #   cache store --namespace billing-service
  #   cache restore --namespace billing-service
  #
  # Internally, this will expand the cache keys, and prepend the value of the
  # namespace.
  #
  # In a Ruby service, running:
  #
  #   cache store --namespace user-service
  #
  # Will expand internally into:
  #
  #   cache store user-service-gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock) vendor/bundle
  #

  cd tests/autocache/ruby
  bundle install --path vendor/bundle

  run cache store --namespace user-service

  assert_success
  assert_output --partial "* Detected Gemfile.lock."
  assert_output --partial "Upload complete."

  rm -rf vendor/bundle

  run cache restore --namespace user-service
  assert_success
  assert_output --partial  "* Fetching 'vendor/bundle' directory with cache keys"
  assert_output --partial "Restored: vendor/bundle/"
}
