#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

PROJECT_ROOT=$(pwd)

setup() {
  eval "$(rbenv init -)"
  source ~/toolbox/sem-version
  source ~/toolbox/sem-install
}

@test "[macOS] sem-version ruby - 2.5.9 " {

  run sem-version ruby 2.5.9
  assert_success
  run ruby --version
  assert_success
  assert_output --partial "2.5.9"
}

@test "[macOS] sem-version ruby - 2.6.7 " {

  run sem-version ruby 2.6.7
  assert_success
  run ruby --version
  assert_success
  assert_output --partial "2.6.7"
}
@test "[macOS] sem-version ruby - 2.7.3 " {

  run sem-version ruby 2.7.3
  assert_success
  run ruby --version
  assert_success
  assert_output --partial "2.7.3"
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

