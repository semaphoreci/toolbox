#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

teardown() {
  rm -rf semaphore-demo-*
}


################################################################################
# cache autostore/autorestore
################################################################################
@test "cache - autostore/autorestore [go]" {

  git clone git@github.com:eddycjy/go-gin-example.git
  cd go-gin-example
  
  go get ./...
  run cache store

  assert_success
  assert_output --partial "* Detected go.sum."
  assert_output --partial "Upload complete."

  sudo rm -rf $HOME/go/*

  run cache restore
  assert_success

  run cache delete go-$SEMAPHORE_GIT_BRANCH-$(checksum go.sum)
  cd ../
  rm -rf go-gin-example
}


@test "cache - autostore/autorestore [bundle]" {

  git clone git@github.com:semaphoreci-demos/semaphore-demo-ruby-rails.git
  cd semaphore-demo-ruby-rails
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
  rm -rf semaphore-demo-ruby-rails
}

@test "cache - autostore/autorestore [pip]" {

  git clone git@github.com:semaphoreci-demos/semaphore-demo-python-django.git
  cd semaphore-demo-python-django
  pip install -r requirements.txt --cache-dir .pip_cache > /dev/null

  run cache store

  assert_success
  assert_output --partial "* Detected requirements.txt"
  assert_output --partial "Upload complete."

  sudo rm -rf .pip_cache

  run cache restore
  assert_success
  assert_output --partial "* Fetching '.pip_cache' directory with cache keys"
  assert_output --partial "Restored: .pip_cache/"

  run cache delete requirements-$SEMAPHORE_GIT_BRANCH-$(checksum requirements.txt)
  cd ../
  rm -rf semaphore-demo-python-django
}

@test "cache - autostore/autorestore [nodejs]" {

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

@test "cache - autostore/autorestore [elixir]" {

  git clone git@github.com:semaphoreci-demos/semaphore-demo-elixir-phoenix.git
  cd semaphore-demo-elixir-phoenix
  mix deps.get > /dev/null

  run cache store

  assert_success
  assert_output --partial "* Detected mix.lock"
  assert_output --partial "Upload complete."

  rm -rf deps

  run cache restore
  assert_success
  assert_output --partial "* Fetching 'deps' directory with cache keys"
  assert_output --partial "Restored: deps/"

  run cache delete deps-$SEMAPHORE_GIT_BRANCH-$(checksum mix.lock)
  cd ../
  rm -rf semaphore-demo-elixir-phoenix
}

@test "cache - autostore/autorestore [php]" {

  git clone git@github.com:semaphoreci-demos/semaphore-demo-php-laravel.git
  cd semaphore-demo-php-laravel
  composer install > /dev/null || true

  run cache store

  assert_success
  assert_output --partial "* Detected composer.lock"
  assert_output --partial "Upload complete."

  sudo rm -rf .pip_cache

  run cache restore
  assert_success
  assert_output --partial "* Fetching 'vendor' directory with cache keys"
  assert_output --partial "Restored: vendor/"

  run cache delete requirements-$SEMAPHORE_GIT_BRANCH-$(checksum composer.lock)
  cd ../
  rm -rf semaphore-demo-php-laravel
}

@test "cache - autostore/autorestore [mvn]" {

  git clone git@github.com:jitpack/maven-simple.git
  cd maven-simple
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

  run cache delete maven-target-$SEMAPHORE_GIT_BRANCH-$(checksum pom.xml)
  run cache delete maven-$SEMAPHORE_GIT_BRANCH-$(checksum pom.xml)
  cd ../
  rm -rf maven-simple
}
