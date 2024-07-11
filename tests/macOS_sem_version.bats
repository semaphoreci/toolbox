#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

PROJECT_ROOT=$(pwd)

setup() {
  eval "$(rbenv init -)"
  source ~/toolbox/sem-version
  source ~/toolbox/sem-install
  source ~/.nvm/nvm.sh
  export NVM_DIR=~/.nvm
}

@test "[macOS] sem-version ruby - 2.7.2 " {

  run sem-version ruby 2.7.2
  assert_success
  run ruby --version
  assert_success
  assert_output --partial "2.7.2"
}
@test "[macOS] sem-version ruby - 3.0.1 " {

  run sem-version ruby 3.0.1
  assert_success
  run ruby --version
  assert_success
  assert_output --partial "3.0.1"
}
@test "[macOS] sem-version php - 8.0.5 " {

  run sem-version php 8.0.5
  assert_failure
}

@test "[macOS] sem-version node - 14.16.1 " {

  run sem-version node 14.16.1
  assert_success
  assert_output --partial "14.16.1"
  node --version
}

