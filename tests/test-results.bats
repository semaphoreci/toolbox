#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  rm -rf /tmp/test-results-cli

  mkdir /tmp/test-results-cli

  artifact yank job test-results/junit.json
  artifact yank job test-results/junit.xml

  cp tests/test-results/junit-sample.xml /tmp/test-results-cli/junit-sample.xml
  cp tests/test-results/junit-sample.json /tmp/test-results-cli/junit-sample.json
}


teardown() {
  rm -rf /tmp/test-results-cli
}

@test "test-results publish works" {
  cd /tmp/test-results-cli

  run test-results publish junit-sample.xml
  assert_success

  run artifact pull job test-results/junit.xml
  assert_success

  run git diff --no-index junit-sample.xml junit.xml --exit-code --output /dev/null
  assert_success

  run artifact pull job test-results/junit.json
  assert_success

  run git diff --no-index junit-sample.json junit.json --exit-code --output /dev/null
  assert_success
}

@test "test-results compile works" {
  cd /tmp/test-results-cli

  run test-results compile junit-sample.xml junit-compile.json
  assert_success
  assert_output --partial "Using rspec parser"

  run git diff --no-index junit-sample.json junit-compile.json --exit-code --output /dev/null
  assert_success
}