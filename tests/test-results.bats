#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  rm -rf /tmp/test-results-cli

  mkdir /tmp/test-results-cli


  cp tests/test-results/junit-sample.xml /tmp/test-results-cli/junit-sample.xml
  cp tests/test-results/junit-sample.json /tmp/test-results-cli/junit-sample.json
}


teardown() {
  rm -rf /tmp/test-results-cli
}

teardown_file() {
  artifact yank job test-results/junit.json
  artifact yank job test-results/junit.xml
}

@test "test-results publish works" {
  cd /tmp/test-results-cli

  run test-results publish junit-sample.xml
  assert_success

  run artifact pull job test-results/junit.xml
  assert_success

  run diff junit-sample.xml junit.xml
  assert_success

  run artifact pull job test-results/junit.json
  assert_success

  run diff junit-sample.json junit.json
  assert_success
}

@test "test-results compile works" {
  cd /tmp/test-results-cli

  run test-results compile junit-sample.xml junit-compile.json
  assert_success
  assert_output --partial "Using rspec parser"

  run diff junit-sample.json junit-compile.json
  assert_success
}