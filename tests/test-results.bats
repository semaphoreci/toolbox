#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  rm -rf /tmp/test-results-cli

  mkdir /tmp/test-results-cli

  cp tests/test-results/junit-sample.xml /tmp/test-results-cli/junit-sample.xml
  cp tests/test-results/junit-sample.json /tmp/test-results-cli/junit-sample.json


  # On macOS, sed requires an extra argument for in-place editing
  if [[ "$OSTYPE" == "darwin"* ]]; then
    export MACOS_EXTRA_ARG="''"
  fi

  # Replace placeholders with actual values
  sed -i $MACOS_EXTRA_ARG "s|JOB_ID|$SEMAPHORE_JOB_ID|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|PPL_ID|$SEMAPHORE_PIPELINE_ID|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|PROJECT_ID|$SEMAPHORE_PROJECT_ID|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|WORKFLOW_ID|$SEMAPHORE_WORKFLOW_ID|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|JOB_NAME|$SEMAPHORE_JOB_NAME|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|JOB_CREATION_TIME|$SEMAPHORE_JOB_CREATION_TIME|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|AGENT_TYPE|$SEMAPHORE_AGENT_MACHINE_TYPE|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|OS_IMAGE|$SEMAPHORE_AGENT_MACHINE_OS_IMAGE|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|GIT_REF_TYPE|$SEMAPHORE_GIT_REF_TYPE|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|GIT_BRANCH|$SEMAPHORE_GIT_BRANCH|g" /tmp/test-results-cli/junit-sample.json
  sed -i $MACOS_EXTRA_ARG "s|GIT_SHA|$SEMAPHORE_GIT_SHA|g" /tmp/test-results-cli/junit-sample.json

  cp tests/test-results/junit-summary.json /tmp/test-results-cli/junit-summary.json
}


teardown() {
  rm -rf /tmp/test-results-cli

  # Clean up artifacts to avoid conflicts in subsequent test runs
  artifact yank job test-results 2>/dev/null || true
  artifact yank workflow test-results/$SEMAPHORE_PIPELINE_ID/$SEMAPHORE_JOB_ID.json 2>/dev/null || true
  artifact yank workflow test-results/$SEMAPHORE_PIPELINE_ID.json 2>/dev/null || true
  artifact yank workflow test-results/$SEMAPHORE_PIPELINE_ID-summary.json 2>/dev/null || true
}

@test "test-results publish works" {
  cd /tmp/test-results-cli

  run test-results publish --no-compress junit-sample.xml
  assert_success

  assert_output --partial "[test-results] Artifact transfers:"
  assert_output --partial "← Pushed: 4 operations"
  assert_output --partial "= Total: 4 operations"

  run artifact pull job test-results/junit.xml
  assert_success

  run diff junit-sample.xml junit.xml
  assert_success

  run artifact pull job test-results/junit.json
  assert_success

  run diff junit-sample.json junit.json
  assert_success

  run artifact pull job test-results/summary.json
  assert_success

  run diff junit-summary.json summary.json
  assert_success
}

@test "test-results compile works" {
  cd /tmp/test-results-cli

  run test-results compile --no-compress junit-sample.xml junit-compile.json
  assert_success
  assert_output --partial "Using rspec parser"

  run diff junit-sample.json junit-compile.json
  assert_success
}

@test "test-results compile does not work with non existent file" {
  cd /tmp/test-results-cli

  run test-results compile --no-compress /tmp/some/file /tmp/some/file.json
  assert_failure
}

@test "test-results publish with --no-raw shows correct operation count" {
  cd /tmp/test-results-cli

  run test-results publish --no-compress --no-raw junit-sample.xml
  assert_success

  assert_output --partial "[test-results] Artifacts pushed: 3 operations"
}

@test "test-results publish multiple files shows correct operation count" {
  cd /tmp/test-results-cli
  cp junit-sample.xml junit-sample2.xml

  run test-results publish --no-compress junit-sample.xml junit-sample2.xml
  assert_success

  assert_output --partial "[test-results] Artifacts pushed: 5 operations"
}

@test "test-results gen-pipeline-report shows transfer summary" {
  cd /tmp/test-results-cli

  run test-results publish --no-compress junit-sample.xml
  assert_success

  mkdir -p pipeline-results
  cp junit-sample.json pipeline-results/

  run test-results gen-pipeline-report --force pipeline-results
  assert_success

  assert_output --partial "[test-results] Artifact transfers:"
  assert_output --partial "← Pushed:"
  assert_output --partial "= Total:"
}
