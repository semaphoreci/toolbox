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
  unset SEMAPHORE_GIT_CLONE_SLOW_GRACE
  unset SEMAPHORE_GIT_CLONE_RETRY_COUNT
  unset SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES
  unset SEMAPHORE_GIT_CLONE_ALT_REGIONS
  unset SEMAPHORE_GIT_CLONE_DOH_CONNECT_TIMEOUT
  unset SEMAPHORE_GIT_CLONE_DOH_MAX_TIME

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

# Install fake curl + dig on PATH so tier 2 resolves alt IPs deterministically
# without network. Returns 185.199.108.133 from DoH, 140.82.121.4 as current IP.
mock_doh_and_dig() {
  local mock_dir="/tmp/slow_mock_net_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<'SCRIPT'
#!/bin/bash
for arg in "$@"; do
  if [[ "$arg" == *"dns.google"* ]]; then
    echo '{"Answer":[{"data":"185.199.108.133"}]}'
    exit 0
  fi
done
exec command curl "$@"
SCRIPT
  chmod +x "$mock_dir/curl"
  cat > "$mock_dir/dig" <<'SCRIPT'
#!/bin/bash
echo "140.82.121.4"
SCRIPT
  chmod +x "$mock_dir/dig"
  export PATH="$mock_dir:$PATH"
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
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0

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

@test "slow retry - speed check grace window protects slow start" {
  # Slow throughput, but the grace window outlasts the process, so it must
  # not be killed as slow (guards against false positives on big-repo startup).
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=999999999
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=5
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=60

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
dd if=/dev/zero of="$dir/.git/objects/pack" bs=1024 count=1 2>/dev/null
sleep 12
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  assert_success
  refute_output --partial "[checkout] Slow clone detected"
}

# === resolve_alt_ips timeout ===

@test "slow retry - resolve_alt_ips passes curl connect/max timeouts" {
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"

  local mock_dir="/tmp/slow_mock_net_$$"
  local args_file="${mock_dir}/curl_args"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<SCRIPT
#!/bin/bash
echo "\$@" >> "${args_file}"
echo '{"Answer":[{"data":"185.199.108.133"}]}'
SCRIPT
  chmod +x "$mock_dir/curl"
  export PATH="$mock_dir:$PATH"

  run checkout::resolve_alt_ips "github.com" "140.82.121.4"
  assert_success
  assert_output --partial "185.199.108.133"

  run cat "${args_file}"
  assert_output --partial "--connect-timeout 5"
  assert_output --partial "--max-time 10"
}

# === resilient_clone integration ===

@test "slow retry - resilient clone retries on slow then reports failure" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=999999999
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=5
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=2
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=0

  mock_doh_and_dig

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
  assert_output --partial "[checkout] Clone failed after 2 attempts"
}

@test "slow retry - resilient clone retries on git error" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=3
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=0

  mock_doh_and_dig

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
  assert_output --partial "[checkout] Clone failed after 3 attempts, trying alternative endpoints"
}

@test "slow retry - resilient clone succeeds on first try when fast" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  refute_output --partial "[checkout] Slow clone detected"
}

@test "slow retry - tier 2 declines for non-github providers" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1
  export SEMAPHORE_GIT_URL="https://gitlab.com/org/repo.git"

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [ "$1" = "clone" ]; then
  echo "fatal: repository not found" >&2
  exit 128
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "Alternative endpoint fallback only supported for github.com (got: gitlab.com)"
}

@test "slow retry - tier 2 declines for SCP-form non-github URLs" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1
  export SEMAPHORE_GIT_URL="git@bitbucket.org:org/repo.git"

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [ "$1" = "clone" ]; then
  echo "fatal: repository not found" >&2
  exit 128
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "Alternative endpoint fallback only supported for github.com (got: bitbucket.org)"
}

# === resolve_alt_ips ===

@test "slow retry - resolve_alt_ips returns mocked IPs different from current" {
  mock_doh_and_dig
  local current_ip="140.82.121.4"

  run checkout::resolve_alt_ips "github.com" "$current_ip"
  assert_success
  [ -n "$output" ]
  refute_output --partial "$current_ip"
  assert_output --partial "185.199.108.133"
}

@test "slow retry - resolve_alt_ips with custom regions" {
  mock_doh_and_dig
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"
  local current_ip="140.82.121.4"

  run checkout::resolve_alt_ips "github.com" "$current_ip"
  assert_success
  assert_output --partial "185.199.108.133"
}

# === clone_with_alt_ip ===

@test "slow retry - clone_with_alt_ip sets ProxyCommand with port 22 for SSH" {
  export SEMAPHORE_GIT_URL="git@github.com:mojombo/grit.git"
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

  run checkout::clone_with_alt_ip "1.2.3.4" "github.com" "22" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "ProxyCommand='nc 1.2.3.4 22'"
}

@test "slow retry - clone_with_alt_ip honors custom SSH port" {
  export SEMAPHORE_GIT_URL="ssh://git@github.com:443/mojombo/grit.git"
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

  run checkout::clone_with_alt_ip "1.2.3.4" "github.com" "443" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "ProxyCommand='nc 1.2.3.4 443'"
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

  run checkout::clone_with_alt_ip "1.2.3.4" "github.com" "22" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "ssh -i /path/to/key -o ProxyCommand='nc 1.2.3.4 22'"

  [ "$GIT_SSH_COMMAND" = "ssh -i /path/to/key" ]
}

@test "slow retry - clone_with_alt_ip uses curloptResolve for HTTPS with port 443" {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=100
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=15

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

  run checkout::clone_with_alt_ip "1.2.3.4" "github.com" "443" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "http.curloptResolve=github.com:443:1.2.3.4"
}

@test "slow retry - clone_with_alt_ip fails cleanly for SSH when nc missing" {
  export SEMAPHORE_GIT_URL="git@github.com:mojombo/grit.git"

  # Restrict PATH to a dir with only a git stub so 'command -v nc' fails.
  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "git should not be called"
SCRIPT
  chmod +x "$mock_dir/git"

  PATH="$mock_dir" run checkout::clone_with_alt_ip "1.2.3.4" "github.com" "22" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "'nc' not available"
  refute_output --partial "git should not be called"
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
