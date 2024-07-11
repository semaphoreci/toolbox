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

@test "[macOS] sem-version ruby - 3.3.3 " {

  run sem-version ruby 3.3.3
  assert_success
  run ruby --version
  assert_success
  assert_output --partial "3.3.3"
}

@test "[macOS] sem-version node - 14.16.1 " {
  run sem-version node 14.16.1
  assert_success
  assert_output --partial "14.16.1"
  run node --version
  assert_success
}

