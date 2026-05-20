#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  unset SEMAPHORE_GIT_REF_TYPE
  unset SEMAPHORE_GIT_TAG_NAME
  unset SEMAPHORE_GIT_PR_SLUG
  unset SEMAPHORE_GIT_PR_NAME
  unset SEMAPHORE_GIT_PR_NUMBER
  unset SEMAPHORE_GIT_CLONE_SLOW_RETRY
  unset SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD
  unset SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT
  unset SEMAPHORE_GIT_CLONE_RETRY_COUNT
  unset SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES
  unset SEMAPHORE_GIT_CLONE_ALT_REGIONS

  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_DIR="repo"
  export SEMAPHORE_GIT_SHA=5608567
  export SEMAPHORE_GIT_REPO_SLUG="mojombo/grit"
  export SEMAPHORE_GIT_REF="refs/heads/master"

  set -u
  source ~/.toolbox/libcheckout
  rm -rf "$SEMAPHORE_GIT_DIR"
  rm -rf /tmp/slow_mock_*
}

teardown() {
  rm -rf "$SEMAPHORE_GIT_DIR"
  rm -rf /tmp/slow_mock_*
}

# === Flag disabled (default) ===

@test "slow retry - disabled by default, normal clone works" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=false
  export SEMAPHORE_GIT_REF_TYPE="push"
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_success
  refute_output --partial "[checkout]"
}

# === Flag enabled, fast clone ===

@test "slow retry - enabled, fast clone succeeds without slow messages" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_REF_TYPE="push"
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_success
  refute_output --partial "[checkout] Slow clone detected"
}

# === clone_with_speed_check unit tests ===

@test "slow retry - speed check succeeds on fast command" {
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

  mkdir -p "${SEMAPHORE_GIT_DIR}"

  run checkout::clone_with_speed_check true
  assert_success
}

@test "slow retry - speed check returns 2 on command error" {
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

  run checkout::clone_with_speed_check false
  [ "$status" -eq 2 ]
}

@test "slow retry - speed check detects and kills slow process" {
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=999999999
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=5

  # Mock: create dir with small data then hang
  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
dd if=/dev/zero of="$dir/.git/objects/pack" bs=1024 count=1 2>/dev/null
sleep 120
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  [ "$status" -eq 1 ]
  assert_output --partial "[checkout] Slow clone detected"
}

# === resilient_clone integration ===

@test "slow retry - resilient clone retries on slow then reports failure" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=999999999
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=5
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=2
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=0

  # Mock git to simulate slow clone
  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [ "$1" = "clone" ]; then
  dir="${@: -1}"
  mkdir -p "$dir/.git/objects"
  dd if=/dev/zero of="$dir/.git/objects/pack" bs=1024 count=1 2>/dev/null
  sleep 120
else
  command git "$@"
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "[checkout] Slow clone detected"
  assert_output --partial "[checkout] Clone still slow after 2 attempts"
}

@test "slow retry - resilient clone retries on git error" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=3

  # Mock git that always fails (not slow, just error)
  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [ "$1" = "clone" ]; then
  echo "fatal: repository not found" >&2
  exit 128
else
  command git "$@"
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "[checkout] Clone failed, retrying..."
  assert_output --partial "[checkout] Clone failed after 3 attempts"
  # Should NOT try alt endpoints for git errors
  refute_output --partial "trying alternative endpoints"
}

@test "slow retry - resilient clone succeeds on first try when fast" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  refute_output --partial "[checkout] Slow clone detected"
}

# === resolve_alt_ips ===

@test "slow retry - resolve_alt_ips returns IPs different from current" {
  local current_ip="140.82.121.35"

  run checkout::resolve_alt_ips "ssh.github.com" "$current_ip"
  assert_success
  # Should return at least one alternative IP
  [ -n "$output" ]
  # Should not contain the current IP
  refute_output --partial "$current_ip"
}

@test "slow retry - resolve_alt_ips with custom regions" {
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"
  local current_ip="140.82.121.35"

  run checkout::resolve_alt_ips "ssh.github.com" "$current_ip"
  assert_success
}

# === clone_with_alt_ip ===

@test "slow retry - clone_with_alt_ip sets ProxyCommand for SSH" {
  export SEMAPHORE_GIT_URL="git@github.com:mojombo/grit.git"
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

  # Mock: capture GIT_SSH_COMMAND instead of cloning
  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "GIT_SSH_COMMAND=$GIT_SSH_COMMAND"
mkdir -p "${@: -1}"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::clone_with_alt_ip "1.2.3.4" "ssh.github.com" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "ProxyCommand='nc 1.2.3.4 22'"
}

@test "slow retry - clone_with_alt_ip preserves existing GIT_SSH_COMMAND" {
  export SEMAPHORE_GIT_URL="git@github.com:mojombo/grit.git"
  export GIT_SSH_COMMAND="ssh -i /path/to/key"
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "GIT_SSH_COMMAND=$GIT_SSH_COMMAND"
mkdir -p "${@: -1}"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::clone_with_alt_ip "1.2.3.4" "ssh.github.com" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "ssh -i /path/to/key -o ProxyCommand='nc 1.2.3.4 22'"

  # GIT_SSH_COMMAND should be restored after
  [ "$GIT_SSH_COMMAND" = "ssh -i /path/to/key" ]
}

@test "slow retry - clone_with_alt_ip uses curloptResolve for HTTPS" {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

  # Mock: capture the -c option
  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "ARGS=$@"
dir="${@: -1}"
mkdir -p "$dir"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::clone_with_alt_ip "1.2.3.4" "github.com" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "http.curloptResolve=github.com:443:1.2.3.4"
}

# === Full checkout flow with slow retry ===

@test "slow retry - full checkout with flag on succeeds for push" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_REF_TYPE="push"
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_success
  assert_output --partial "HEAD is now at 91940c2"
  refute_output --partial "[checkout] Slow clone detected"
}

@test "slow retry - full checkout with flag on succeeds for PR" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_REF_TYPE="pull-request"
  export SEMAPHORE_GIT_REF="refs/pull/186/merge"
  export SEMAPHORE_GIT_SHA=30774365e11f2b1e18706c9ed0920369f6d7c205

  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA"
}

@test "slow retry - full checkout with flag on succeeds for tag" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_REF_TYPE="tag"
  export SEMAPHORE_GIT_TAG_NAME='v2.4.1'
  export SEMAPHORE_GIT_SHA=91940c2cc18ec08b751482f806f1b8bfa03d98a5

  run checkout
  assert_success
  assert_output --partial "HEAD is now at $SEMAPHORE_GIT_SHA Release $SEMAPHORE_GIT_TAG_NAME"
}
