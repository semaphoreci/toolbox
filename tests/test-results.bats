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

  assert_output --partial "[test-results] Artifact transfers:"
  assert_output --partial "← Pushed: 3 operations"
  assert_output --partial "= Total: 3 operations"
}

@test "test-results publish multiple files shows correct operation count" {
  cd /tmp/test-results-cli
  cp junit-sample.xml junit-sample2.xml

  run test-results publish --no-compress junit-sample.xml junit-sample2.xml
  assert_success

  assert_output --partial "[test-results] Artifact transfers:"
  assert_output --partial "← Pushed: 5 operations"
  assert_output --partial "= Total: 5 operations"
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

@test "test-results compile with file:parser syntax for single file" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .

  run test-results compile --no-compress golang.xml:junit output.json
  assert_success
  assert_output --partial "Using junit parser"
  
  [ -f output.json ]
}

@test "test-results compile with file:parser syntax for multiple files with different parsers" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .
  cp $BATS_TEST_DIRNAME/test-results/staticcheck.json .

  run test-results compile --no-compress golang.xml:junit staticcheck.json:go:staticcheck output.json
  assert_success
  
  assert_output --partial "Using junit parser"
  assert_output --partial "Using go:staticcheck parser"
  
  [ -f output.json ]
}

@test "test-results compile with mix of explicit parser and auto-detect" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/junit-sample.xml .
  cp $BATS_TEST_DIRNAME/test-results/staticcheck.json .

  run test-results compile --no-compress junit-sample.xml staticcheck.json:go:staticcheck output.json
  assert_success
  
  assert_output --partial "Using rspec parser"
  assert_output --partial "Using go:staticcheck parser"
  
  [ -f output.json ]
}

@test "test-results publish with file:parser syntax for single file" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .

  run test-results publish --no-compress golang.xml:junit
  assert_success
  assert_output --partial "Using junit parser"
  
  assert_output --partial "[test-results] Artifact transfers:"
  assert_output --partial "← Pushed:"
}

@test "test-results publish with file:parser syntax for multiple files" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .
  cp $BATS_TEST_DIRNAME/test-results/revive.json .

  run test-results publish --no-compress golang.xml:junit revive.json:go:revive
  assert_success
  
  assert_output --partial "Using junit parser"
  assert_output --partial "Using go:revive parser"
  assert_output --partial "← Pushed: 5 operations"
}

@test "test-results compile fails without --ignore-missing when file doesn't exist" {
  cd /tmp/test-results-cli

  run test-results compile --no-compress missing-file.xml output.json
  assert_failure
  assert_output --partial "failed to stat missing-file.xml"
}

@test "test-results compile succeeds with --ignore-missing when file doesn't exist" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/junit-sample.xml .

  run test-results compile --no-compress --ignore-missing junit-sample.xml missing-file.xml output.json
  assert_success
  assert_output --partial "File not found, skipping: missing-file.xml"
  assert_output --partial "Using rspec parser"
  
  [ -f output.json ]
}

@test "test-results compile with --ignore-missing processes only existing files" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .
  cp $BATS_TEST_DIRNAME/test-results/rspec2.xml .

  run test-results compile --no-compress --ignore-missing golang.xml missing1.xml rspec2.xml missing2.xml output.json
  assert_success
  
  assert_output --partial "File not found, skipping: missing1.xml"
  assert_output --partial "File not found, skipping: missing2.xml"
  assert_output --partial "Using junit parser"
  
  [ -f output.json ]
}

@test "test-results publish fails without --ignore-missing when file doesn't exist" {
  cd /tmp/test-results-cli

  run test-results publish --no-compress missing-file.xml
  assert_failure
  assert_output --partial "failed to stat missing-file.xml"
}

@test "test-results publish succeeds with --ignore-missing when file doesn't exist" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/junit-sample.xml .

  run test-results publish --no-compress --ignore-missing junit-sample.xml missing-file.xml
  assert_success
  assert_output --partial "File not found, skipping: missing-file.xml"
  assert_output --partial "Using rspec parser"
  assert_output --partial "[test-results] Artifact transfers:"
}

# Edge cases

@test "test-results compile file:parser overrides global --parser flag" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .

  run test-results compile --no-compress --parser rspec golang.xml:junit output.json
  assert_success
  
  assert_output --partial "Using junit parser"
  refute_output --partial "Using rspec parser"
  
  [ -f output.json ]
}

@test "test-results compile with file:parser and --ignore-missing combined" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .
  cp $BATS_TEST_DIRNAME/test-results/staticcheck.json .

  run test-results compile --no-compress --ignore-missing golang.xml:junit missing.xml:rspec staticcheck.json:go:staticcheck output.json
  assert_success
  
  assert_output --partial "File not found, skipping: missing.xml"
  assert_output --partial "Using junit parser"
  assert_output --partial "Using go:staticcheck parser"
  
  [ -f output.json ]
}

@test "test-results publish with invalid parser in file:parser syntax" {
  cd /tmp/test-results-cli
  cp $BATS_TEST_DIRNAME/test-results/golang.xml .

  run test-results publish --no-compress golang.xml:invalid-parser
  assert_failure
  assert_output --partial "parser not found: invalid-parser"
}

@test "test-results compile creates empty result when all files are missing with --ignore-missing" {
  cd /tmp/test-results-cli

  run test-results compile --no-compress --ignore-missing missing1.xml missing2.json output.json
  assert_success
  assert_output --partial "No files to process"
  
  [ -f output.json ]
}
