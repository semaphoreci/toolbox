#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

PROJECT_ROOT=$(pwd)

setup() {
  cache clear
  cd $PROJECT_ROOT
}

@test "cache - autostore/autorestore [go]" {
  run cache delete go-$SEMAPHORE_GIT_BRANCH-$(checksum go.sum)
  cd tests/autocache/go
  go get ./...

  run cache store

  assert_success
  assert_output --partial "* Detected go.sum."
  assert_output --partial "Upload complete."

  sudo rm -rf $HOME/go/*

  run cache restore
  assert_success
}

@test "cache - autostore/autorestore [bundle]" {
  run cache delete gems-$SEMAPHORE_GIT_BRANCH-$(checksum Gemfile.lock)
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

@test "cache - autostore/autorestore [pip]" {
  run cache delete requirements-$SEMAPHORE_GIT_BRANCH-$(checksum requirements.txt)
  cd tests/autocache/python
  pip install -r requirements.txt --cache-dir .pip_cache

  run cache store

  assert_success
  assert_output --partial "* Detected requirements.txt"
  assert_output --partial "Upload complete."

  sudo rm -rf .pip_cache

  run cache restore
  assert_success
  assert_output --partial "* Fetching '.pip_cache' directory with cache keys"
  assert_output --partial "Restored: .pip_cache/"

}

@test "cache - autostore/autorestore [nodejs]" {
  run cache delete node-modules-$SEMAPHORE_GIT_BRANCH-$(checksum package-lock.json)
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

@test "cache - autostore/autorestore [elixir]" {
  run cache delete deps-$SEMAPHORE_GIT_BRANCH-$(checksum mix.lock)
  cd tests/autocache/elixir
  mix deps.get

  run cache store

  assert_success
  assert_output --partial "* Detected mix.lock"
  assert_output --partial "Upload complete."

  rm -rf deps

  run cache restore
  assert_success
  assert_output --partial "* Fetching 'deps' directory with cache keys"
  assert_output --partial "Restored: deps/"
}

@test "cache - autostore/autorestore [php]" {
  run cache delete requirements-$SEMAPHORE_GIT_BRANCH-$(checksum composer.lock)
  cd tests/autocache/php
  composer install

  run cache store

  assert_success
  assert_output --partial "* Detected composer.lock"
  assert_output --partial "Upload complete."

  sudo rm -rf .pip_cache

  run cache restore
  assert_success
  assert_output --partial "* Fetching 'vendor' directory with cache keys"
  assert_output --partial "Restored: vendor/"
}

@test "cache - autostore/autorestore [mvn]" {
  run cache delete maven-target-$SEMAPHORE_GIT_BRANCH-$(checksum pom.xml)
  run cache delete maven-$SEMAPHORE_GIT_BRANCH-$(checksum pom.xml)

  cd tests/autocache/java
  mvn -Dmaven.repo.local=".m2" test-compile

  run cache store

  assert_success
  assert_output --partial "* Detected pom.xml"
  assert_output --partial "* Using default cache path '.m2'."
  assert_output --partial "Upload complete."
  assert_output --partial "* Using default cache path 'target'."
  assert_output --partial "Upload complete."

  sudo rm -rf .m2 target

  run cache restore
  assert_success
  assert_output --partial "* Fetching '.m2' directory with cache keys"
  assert_output --partial "Restored: .m2"
  assert_output --partial "* Fetching 'target' directory with cache keys"
  assert_output --partial "Restored: target"
}
