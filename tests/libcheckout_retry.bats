#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

# Retry behaviour of checkout::git_clone / checkout::git_fetch, without network:
# the remote is a local bare repository and a fake `git` on the PATH fails the
# first N calls of a given subcommand with a chosen message before handing over
# to the real git.

setup() {
  unset SEMAPHORE_GIT_REF_TYPE
  unset SEMAPHORE_GIT_TAG_NAME
  unset SEMAPHORE_GIT_PARTIAL_CLONE_FILTER
  unset SEMAPHORE_GIT_SPARSE_CHECKOUT_PATHS
  unset SEMAPHORE_GIT_CLONE_SLOW_RETRY
  unset SEMAPHORE_GIT_RETRY_ATTEMPTS

  export SEMAPHORE_GIT_RETRY_BACKOFF=0
  export SEMAPHORE_GIT_DIR="repo"
  export SEMAPHORE_GIT_BRANCH=main
  export SEMAPHORE_GIT_REF="refs/heads/main"

  export RETRY_TEST_DIR
  RETRY_TEST_DIR="$(mktemp -d)"
  make_remote
  export SEMAPHORE_GIT_URL="file://${RETRY_TEST_DIR}/remote.git"

  set -u
  source ~/.toolbox/libcheckout
  rm -rf "$SEMAPHORE_GIT_DIR"
}

teardown() {
  rm -rf "$SEMAPHORE_GIT_DIR"
  rm -rf "$RETRY_TEST_DIR"
}

make_remote() {
  local work="${RETRY_TEST_DIR}/work"
  git init -q --initial-branch main "$work"
  git -C "$work" -c user.name=Ada -c user.email=ada@example.com commit -q --allow-empty -m "first"
  git -C "$work" -c user.name=Ada -c user.email=ada@example.com commit -q --allow-empty -m "second"
  git clone -q --bare "$work" "${RETRY_TEST_DIR}/remote.git"
  export SEMAPHORE_GIT_SHA
  SEMAPHORE_GIT_SHA="$(git -C "$work" rev-parse HEAD)"
}

# fake_git <subcommand> <failures> <message>: the first <failures> invocations of
# `git <subcommand>` print <message> on stderr and exit 128; later ones (and every
# other subcommand) run the real git.
fake_git() {
  local bin="${RETRY_TEST_DIR}/bin"
  mkdir -p "$bin"
  echo 0 > "${RETRY_TEST_DIR}/calls"
  cat > "$bin/git" <<SCRIPT
#!/bin/bash
real_git="$(command -v git)"
if [ "\$1" = "$1" ]; then
  calls=\$(cat "${RETRY_TEST_DIR}/calls")
  calls=\$((calls + 1))
  echo "\$calls" > "${RETRY_TEST_DIR}/calls"
  if [ "\$calls" -le "$2" ]; then
    echo "$3" >&2
    exit 128
  fi
fi
exec "\$real_git" "\$@"
SCRIPT
  chmod +x "$bin/git"
  export PATH="$bin:$PATH"
}

calls() { cat "${RETRY_TEST_DIR}/calls"; }

@test "retry - a clone that works first time leaves no trace" {
  run checkout
  assert_success
  assert_output --partial "HEAD is now at ${SEMAPHORE_GIT_SHA:0:7}"
  refute_output --partial "[checkout]"
}

@test "retry - a transient ssh error is retried until the clone succeeds" {
  fake_git clone 2 "kex_exchange_identification: Connection reset by peer"

  run checkout
  assert_success
  assert_output --partial "'git clone' failed (exit 128), retrying in 0s (attempt 1/5)"
  assert_output --partial "'git clone' failed (exit 128), retrying in 0s (attempt 2/5)"
  assert_output --partial "HEAD is now at ${SEMAPHORE_GIT_SHA:0:7}"
  assert_equal "$(calls)" 3
}

@test "retry - the target directory is wiped between attempts" {
  # A leftover from a failed attempt must not turn the retry into
  # "destination path already exists".
  fake_git clone 1 "error: failed to fetch some objects from 'https://example.com/info/lfs'"
  mkdir -p "$SEMAPHORE_GIT_DIR/stale"

  run checkout
  assert_success
  refute_output --partial "already exists"
  assert_output --partial "HEAD is now at ${SEMAPHORE_GIT_SHA:0:7}"
}

@test "retry - gives up after SEMAPHORE_GIT_RETRY_ATTEMPTS and shows the last error" {
  export SEMAPHORE_GIT_RETRY_ATTEMPTS=3
  fake_git clone 10 "fatal: early EOF"

  run checkout
  assert_failure
  assert_output --partial "'git clone' failed after 3 attempts: fatal: early EOF"
  refute_output --partial "Branch not found performing full clone"
  assert_equal "$(calls)" 3
}

@test "retry - a missing repository is not retried" {
  fake_git clone 10 "ERROR: Repository not found."

  run checkout
  assert_failure
  assert_output --partial "'git clone' failed and retrying would not help: ERROR: Repository not found."
  refute_output --partial "Branch not found performing full clone"
  assert_equal "$(calls)" 1
}

@test "retry - a refused deploy key is not retried" {
  fake_git clone 10 "git@example.com: Permission denied (publickey)."

  run checkout
  assert_failure
  assert_output --partial "retrying would not help: git@example.com: Permission denied (publickey)."
  assert_equal "$(calls)" 1
}

@test "retry - a missing LFS object is reported instead of 'Branch not found'" {
  fake_git clone 10 "Object does not exist on the server: [404] Object does not exist on the server"

  run checkout
  assert_failure
  assert_output --partial "retrying would not help: Object does not exist on the server"
  refute_output --partial "Branch not found performing full clone"
  refute_output --partial "Branch: ${SEMAPHORE_GIT_BRANCH} not found"
  assert_equal "$(calls)" 1
}

@test "retry - a missing branch still falls back to a full clone" {
  export SEMAPHORE_GIT_BRANCH=nope

  run checkout
  assert_failure
  assert_output --partial "retrying would not help: fatal: Remote branch nope not found in upstream origin"
  assert_output --partial "Branch not found performing full clone"
  assert_output --partial "Branch: nope not found .... Exiting"
}

@test "retry - a transient fetch error is retried (pull request ref)" {
  export SEMAPHORE_GIT_REF_TYPE="pull-request"
  export SEMAPHORE_GIT_REF="refs/heads/main"
  fake_git fetch 1 "ssh_exchange_identification: read: Connection reset by peer"

  run checkout
  assert_success
  assert_output --partial "'git fetch' failed (exit 128), retrying in 0s (attempt 1/5)"
  assert_output --partial "HEAD is now at ${SEMAPHORE_GIT_SHA:0:7}"
}

@test "retry - a fetch error is no longer hidden" {
  export SEMAPHORE_GIT_REF_TYPE="pull-request"
  export SEMAPHORE_GIT_REF="refs/pull/1/head"

  run checkout
  assert_failure
  assert_output --partial "couldn't find remote ref refs/pull/1/head"
  assert_output --partial "Revision: ${SEMAPHORE_GIT_SHA} not found .... Exiting"
}

@test "retry - the slow-retry opt-in path is untouched" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1
  fake_git clone 1 "fatal: early EOF"

  run checkout
  assert_failure
  assert_output --partial "[checkout] Clone failed after 1 attempts"
  refute_output --partial "retrying in"
}
