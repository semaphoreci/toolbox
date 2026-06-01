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
  unset SEMAPHORE_GIT_CLONE_STALL_TIMEOUT
  unset SEMAPHORE_GIT_CLONE_SSH_STALL_TIMEOUT
  unset SEMAPHORE_GIT_CLONE_OVERALL_TIMEOUT
  unset SEMAPHORE_GIT_CLONE_LOW_SPEED_LIMIT
  unset SEMAPHORE_GIT_CLONE_LOW_SPEED_TIME
  unset SEMAPHORE_GIT_CLONE_SLOW_GRACE
  unset SEMAPHORE_GIT_CLONE_CHECK_INTERVAL
  unset SEMAPHORE_GIT_CLONE_RETRY_COUNT
  unset SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES
  unset SEMAPHORE_GIT_CLONE_ALT_REGIONS
  unset SEMAPHORE_GIT_CLONE_DOH_PROVIDERS
  unset SEMAPHORE_GIT_CLONE_DOH_CONNECT_TIMEOUT
  unset SEMAPHORE_GIT_CLONE_DOH_MAX_TIME
  unset CHECKOUT_LAST_CLONE_CLASS
  unset CHECKOUT_CLONE_ERRLOG

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
  rm -f /tmp/pwned
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

@test "slow retry - speed check detects and kills a mid-clone stall" {
  # Writes a little, then makes no further progress: the stall watchdog must
  # abort it. No throughput floor is configured (the safe default).
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=2
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
dd if=/dev/zero of="$dir/.git/objects/pack" bs=4096 count=4 2>/dev/null
sleep 120
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  [ "$status" -eq 1 ]
  assert_output --partial "[checkout] Clone stalled: no progress"
}

@test "slow retry - speed check aborts a pre-write hang (size never leaves 0)" {
  # The clone hangs before ever writing the target dir. The old cur_size>0
  # gate meant the watchdog never armed; the stall watchdog must still fire.
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=2
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
# Never creates "$1" (the target dir); just hangs.
sleep 120
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  [ "$status" -eq 1 ]
  assert_output --partial "[checkout] Clone stalled: no progress"
}

@test "slow retry - speed check does NOT kill a slow-but-progressing clone" {
  # Grows only ~8KB/s — well under the old 20000 B/s default threshold that
  # would have killed it — but it makes continuous forward progress, so the
  # stall watchdog (the safe default, throughput floor off) must let it finish.
  # This is the core false-positive regression from review finding #1.
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=4
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=2
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1
  # throughput floor left at its default (0 = disabled)

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
for _ in $(seq 1 8); do
  dd if=/dev/zero bs=1024 count=8 >> "$dir/.git/objects/pack" 2>/dev/null
  sleep 1
done
exit 0
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  assert_success
  refute_output --partial "stalled"
  refute_output --partial "Slow clone detected"
}

@test "slow retry - BY DEFAULT a long zero-growth (CPU) phase is NOT killed" {
  # Critical regression (review C1 / R2#2): with no opt-in du backstop set, the
  # watchdog must not kill a clone whose on-disk size plateaus during a CPU-only
  # phase (resolving deltas / checkout). Writes data fast, then idles with zero
  # growth far longer than any old default stall window, then exits 0.
  # No SEMAPHORE_GIT_CLONE_STALL_TIMEOUT / _OVERALL_TIMEOUT is set on purpose.
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
dd if=/dev/zero of="$dir/.git/objects/pack" bs=1024 count=512 2>/dev/null  # ~5MB fast
sleep 10   # long CPU-only plateau, no growth
exit 0
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  assert_success
  refute_output --partial "stalled"
  refute_output --partial "wall-clock"
}

@test "slow retry - stall guard arms only AFTER the grace window" {
  # Writes during the grace window, then hangs with zero further progress. The
  # write inside grace must not immunise the later hang: detection arms once
  # grace elapses and the stall guard then fires. (Boundary test for the grace
  # window — the kill comes from the watchdog, not the process exiting.)
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=4
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=3
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
dd if=/dev/zero of="$dir/.git/objects/pack" bs=4096 count=4 2>/dev/null
sleep 120
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  [ "$status" -eq 1 ]
  assert_output --partial "[checkout] Clone stalled: no progress"
}

@test "slow retry - absolute wall-clock cap aborts a clone that never exits" {
  # Opt-in backstop: even a clone that keeps growing forever is abandoned once
  # it blows the wall-clock cap.
  export SEMAPHORE_GIT_CLONE_OVERALL_TIMEOUT=3
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=100
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
for _ in $(seq 1 60); do
  dd if=/dev/zero bs=4096 count=8 >> "$dir/.git/objects/pack" 2>/dev/null
  sleep 1
done
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  [ "$status" -eq 1 ]
  assert_output --partial "[checkout] Clone exceeded wall-clock cap"
}

@test "slow retry - stall kill reaps the whole process group (no orphans)" {
  # The clone spawns a long-lived child and then hangs. A parent-only kill
  # would orphan the child; the process-group kill must reap it. (#3)
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=2
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1
  export GRANDCHILD_PIDFILE="/tmp/slow_mock_gchild_$$"
  rm -f "$GRANDCHILD_PIDFILE"

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir/.git/objects"
dd if=/dev/zero of="$dir/.git/objects/pack" bs=4096 count=4 2>/dev/null
# Long-lived grandchild with detached stdio so it cannot hold the test's
# output pipe open; we assert it is killed via the process group.
sleep 300 >/dev/null 2>&1 &
echo $! > "$GRANDCHILD_PIDFILE"
sleep 300
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  [ "$status" -eq 1 ]

  local gpid
  gpid="$(cat "$GRANDCHILD_PIDFILE")"
  [ -n "$gpid" ]
  # Give signals a moment to propagate through the group.
  sleep 1
  run kill -0 "$gpid"
  assert_failure
}

@test "slow retry - speed check grace window protects slow start" {
  # No on-disk growth, but the grace window outlasts the process, so the stall
  # guard must never arm (guards against false positives on big-repo startup).
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=5
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
  refute_output --partial "stalled"
}

# === resolve_alt_ips timeout ===

@test "slow retry - resolve_alt_ips passes curl connect/max timeouts" {
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"

  local mock_dir="/tmp/slow_mock_net_$$"
  local args_file="${mock_dir}/curl_args"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<SCRIPT
#!/bin/bash
printf '%s\n' "\$@" >> "${args_file}"
echo '{"Answer":[{"data":"185.199.108.133"}]}'
SCRIPT
  chmod +x "$mock_dir/curl"
  export PATH="$mock_dir:$PATH"

  run checkout::resolve_alt_ips "github.com" "140.82.121.4"
  assert_success
  assert_output --partial "185.199.108.133"

  # printf put each arg on its own line; assert flag and value both reached curl.
  run cat "${args_file}"
  assert_line "--connect-timeout"
  assert_line "5"
  assert_line "--max-time"
  assert_line "10"
}

# === http.lowSpeedLimit / lowSpeedTime injection ===

@test "slow retry - injects http.lowSpeedLimit/Time for HTTPS clones" {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_CLONE_LOW_SPEED_LIMIT=2048
  export SEMAPHORE_GIT_CLONE_LOW_SPEED_TIME=25

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "ARGS=$*"
mkdir -p "${@: -1}"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::clone_with_speed_check git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "http.lowSpeedLimit=2048"
  assert_output --partial "http.lowSpeedTime=25"
}

@test "slow retry - does NOT inject low-speed config for SSH clones" {
  export SEMAPHORE_GIT_URL="git@github.com:mojombo/grit.git"

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "ARGS=$*"
mkdir -p "${@: -1}"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::clone_with_speed_check git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  refute_output --partial "http.lowSpeedLimit"
}

# === resilient_clone integration ===

@test "slow retry - resilient clone retries on stall then reports failure" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=2
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=2
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=0

  mock_doh_and_dig

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
# Robust to injected global options (e.g. -c http.lowSpeedLimit=...): match the
# clone subcommand wherever it appears.
if [[ " $* " == *" clone "* ]]; then
  dir="${@: -1}"
  mkdir -p "$dir/.git/objects"
  dd if=/dev/zero of="$dir/.git/objects/pack" bs=4096 count=4 2>/dev/null
  sleep 120
else
  command git "$@"
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "[checkout] Clone stalled"
  assert_output --partial "[checkout] Clone failed after 2 attempts"
}

@test "slow retry - resilient clone retries a transient (network) git error 3x by default" {
  # No explicit retry count -> must default to 3 (matching legacy `retry`). A
  # network-class git error is transient, so it is retried and then escalates
  # to the alt-endpoint tier. (#4)
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=0

  mock_doh_and_dig

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [[ " $* " == *" clone "* ]]; then
  echo "fatal: unable to access 'https://github.com/...': Could not resolve host: github.com" >&2
  exit 128
else
  command git "$@"
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "[checkout] Clone failed (network), retrying... (attempt 2/3)"
  assert_output --partial "[checkout] Clone failed after 3 attempts, trying alternative endpoints"
}

@test "slow retry - resilient clone does NOT retry or try alt endpoints on a definitive git error" {
  # A repo-not-found / auth style failure cannot be fixed by retrying or by
  # switching endpoints. It must fail fast: no retries, no DoH, no tier 2. (#7)
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=3
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=3

  # Fail loudly if DoH is ever consulted for a definitive error.
  local mock_dir="/tmp/slow_mock_net_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<SCRIPT
#!/bin/bash
echo "called" >> "${mock_dir}/curl_called"
echo '{"Answer":[{"data":"185.199.108.133"}]}'
SCRIPT
  chmod +x "$mock_dir/curl"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [[ " $* " == *" clone "* ]]; then
  echo "fatal: repository not found" >&2
  exit 128
else
  command git "$@"
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  [ "$status" -eq 3 ]
  assert_output --partial "definitive error (notfound-repo)"
  assert_output --partial "event=definitive_error"
  refute_output --partial "retrying..."
  refute_output --partial "trying alternative endpoints"
  refute_output --partial "Switching to alternative"
  refute_output --partial "event=tier2_enter"
  [ ! -f "${mock_dir}/curl_called" ]
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
if [[ " $* " == *" clone "* ]]; then
  echo "fatal: unable to access: Could not resolve host: gitlab.com" >&2
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
if [[ " $* " == *" clone "* ]]; then
  echo "fatal: unable to access: Could not resolve host: bitbucket.org" >&2
  exit 128
fi
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "Alternative endpoint fallback only supported for github.com (got: bitbucket.org)"
}

@test "slow retry - tier 2 is eligible for ssh.github.com (SSH over 443)" {
  # ssh://git@ssh.github.com:443/... must NOT be excluded from the tier-2 gate.
  # A transient failure should reach alt-endpoint resolution. (#6)
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=1
  export SEMAPHORE_GIT_URL="ssh://git@ssh.github.com:443/mojombo/grit.git"

  mock_doh_and_dig

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [[ " $* " == *" clone "* ]]; then
  echo "fatal: Could not read from remote repository (connection timed out)" >&2
  exit 128
fi
command git "$@"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  refute_output --partial "only supported for github.com"
  assert_output --partial "Switching to alternative GitHub endpoint (185.199.108.133)"
}

@test "slow retry - tier 2 eligible for HTTPS github.com with explicit port and userinfo" {
  # https://github.com:443/... and https://user@github.com/... must normalize to
  # the bare host and remain tier-2 eligible (R2-Med host-parsing finding).
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=0

  mock_doh_and_dig

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [[ " $* " == *" clone "* ]]; then
  echo "fatal: unable to access: Could not resolve host: github.com" >&2
  exit 128
fi
command git "$@"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  local url
  for url in "https://github.com:443/mojombo/grit.git" "https://git@github.com/mojombo/grit.git"; do
    export SEMAPHORE_GIT_URL="$url"
    run checkout::resilient_clone "$url" "${SEMAPHORE_GIT_DIR}"
    refute_output --partial "only supported for github.com"
    # Reached tier 2 and tried alt-endpoint resolution rather than declining.
    assert_output --partial "trying alternative endpoints"
  done
}

@test "slow retry - tier 2 recovers a clone via an alternative IP (end-to-end)" {
  # The headline value proposition: tier 1 fails on the primary path, then the
  # clone SUCCEEDS through a DoH-resolved alternative IP. The mock fails the
  # direct attempt and succeeds only when routed via http.curloptResolve to the
  # alt IP, proving failover actually recovers a clone.
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=2
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"

  mock_doh_and_dig

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [[ " $* " == *" clone "* ]]; then
  if [[ "$*" == *"curloptResolve"*"185.199.108.133"* ]]; then
    # Routed through the alternative GitHub IP -> succeed.
    dir="${@: -1}"
    mkdir -p "$dir/.git"
    exit 0
  fi
  # Direct attempt against the primary endpoint -> transient network failure.
  echo "fatal: unable to access: Could not resolve host: github.com" >&2
  exit 128
fi
command git "$@"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::resilient_clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "Switching to alternative GitHub endpoint (185.199.108.133)"
  assert_output --partial "event=tier2_success"
  refute_output --partial "All clone attempts failed"
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

# === Error classification ===

classify_fixture() {
  local f
  f="$(mktemp)"
  printf '%s\n' "$1" > "$f"
  run checkout::classify_clone_error "$f"
  rm -f "$f"
}

@test "classify_clone_error - missing branch" {
  classify_fixture "fatal: Remote branch nope not found in upstream origin"
  assert_output "notfound-branch"
}

@test "classify_clone_error - bare 'not found in upstream' is a missing branch, not repo" {
  # No literal "remote branch" prefix; must still classify as a missing branch
  # so checkout::shallow escalates to a full clone (review round-3 R1 finding).
  classify_fixture "fatal: branch 'feature/x' not found in upstream origin"
  assert_output "notfound-branch"
}

@test "classify_clone_error - shallow capability unsupported" {
  classify_fixture "fatal: dumb http transport does not support shallow capabilities"
  assert_output "shallow-unsupported"
}

@test "classify_clone_error - missing repository" {
  classify_fixture "remote: Repository not found.
fatal: repository 'https://github.com/x/y.git/' not found"
  assert_output "notfound-repo"
}

@test "classify_clone_error - authentication failure" {
  classify_fixture "remote: Invalid username or password.
fatal: Authentication failed for 'https://github.com/x/y.git/'"
  assert_output "auth"
}

@test "classify_clone_error - permission denied (ssh)" {
  classify_fixture "git@github.com: Permission denied (publickey)."
  assert_output "auth"
}

@test "classify_clone_error - network: could not resolve host" {
  classify_fixture "fatal: unable to access: Could not resolve host: github.com"
  assert_output "network"
}

@test "classify_clone_error - network: connection timed out" {
  classify_fixture "ssh: connect to host ssh.github.com port 443: Connection timed out"
  assert_output "network"
}

@test "classify_clone_error - network: early EOF / rpc failed" {
  classify_fixture "error: RPC failed; curl 18 transfer closed
fatal: early EOF"
  assert_output "network"
}

@test "classify_clone_error - unknown for unrelated text" {
  classify_fixture "warning: something benign happened"
  assert_output "unknown"
}

@test "classify_clone_error - unknown for empty/missing log" {
  run checkout::classify_clone_error ""
  assert_output "unknown"
  run checkout::classify_clone_error "/nonexistent/path/$$"
  assert_output "unknown"
}

# === Full-clone fallback decision (review finding #5) ===

@test "should_full_clone_fallback - resilient transient (rc=1) does not escalate" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  CHECKOUT_LAST_CLONE_CLASS="stall"
  run checkout::should_full_clone_fallback 1
  assert_failure
}

@test "should_full_clone_fallback - resilient branch-not-found (rc=3) escalates" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  CHECKOUT_LAST_CLONE_CLASS="notfound-branch"
  run checkout::should_full_clone_fallback 3
  assert_success
}

@test "should_full_clone_fallback - resilient auth (rc=3) does not escalate" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  CHECKOUT_LAST_CLONE_CLASS="auth"
  run checkout::should_full_clone_fallback 3
  assert_failure
}

@test "should_full_clone_fallback - resilient repo-not-found (rc=3) does not escalate" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  CHECKOUT_LAST_CLONE_CLASS="notfound-repo"
  run checkout::should_full_clone_fallback 3
  assert_failure
}

@test "should_full_clone_fallback - shallow-unsupported (rc=3) escalates to full clone" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  CHECKOUT_LAST_CLONE_CLASS="shallow-unsupported"
  run checkout::should_full_clone_fallback 3
  assert_success
}

@test "should_full_clone_fallback - legacy path always escalates" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=false
  run checkout::should_full_clone_fallback 1
  assert_success
}

# === Shallow clone: full-history escalation is gated (review finding #5) ===

@test "shallow - transient failure does NOT escalate to a full-history clone" {
  # The shallow (-b branch) attempt fails for a network reason. Shallow must
  # NOT silently download full history; it must propagate the failure.
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1
  export SEMAPHORE_GIT_URL="https://gitlab.com/org/repo.git"
  export CLONE_LOG="/tmp/slow_mock_clonelog_$$"
  rm -f "$CLONE_LOG"

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [[ " $* " == *" clone "* ]]; then
  echo "$@" >> "$CLONE_LOG"
  echo "fatal: unable to access: Could not resolve host: gitlab.com" >&2
  exit 128
fi
command git "$@"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::shallow
  assert_failure
  assert_output --partial "not falling back to full-history clone"
  refute_output --partial "Branch not found performing full clone"

  # Exactly one clone attempt (the shallow one, carrying -b); no full clone.
  run cat "$CLONE_LOG"
  assert_output --partial "-b master"
  [ "$(grep -c 'clone' "$CLONE_LOG")" -eq 1 ]
}

@test "shallow - genuine missing branch DOES escalate to a full-history clone" {
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=2

  # Fully self-contained git mock: it handles every subcommand so no real git
  # ever runs (a real `git checkout` here could escape into the mounted
  # workspace repo). The shallow attempt (carrying -b) reports a missing
  # branch; the full-clone fallback and all post-clone steps succeed.
  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
# Robust to injected global -c options before the subcommand.
if [[ " $* " == *" clone "* ]]; then
  if [[ " $* " == *" -b "* ]]; then
    echo "fatal: Remote branch master not found in upstream origin" >&2
    exit 128
  fi
  mkdir -p "${@: -1}/.git"
  exit 0
fi
# checkout / rev-list / reset / show-ref / etc. -> pretend success.
exit 0
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::shallow
  assert_output --partial "Branch not found performing full clone"
  refute_output --partial "not falling back to full-history clone"
}

@test "shallow - server without shallow support DOES escalate to a full clone" {
  # 'dumb http transport does not support shallow capabilities' is recoverable
  # by a non-shallow (full) clone. It must escalate, not fail hard. (R2-High)
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true
  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=2

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
if [[ " $* " == *" clone "* ]]; then
  if [[ " $* " == *" --depth "* ]]; then
    # The shallow attempt is rejected by the server/proxy.
    echo "fatal: dumb http transport does not support shallow capabilities" >&2
    exit 128
  fi
  # The full (non-shallow) clone succeeds.
  mkdir -p "${@: -1}/.git"
  exit 0
fi
exit 0
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::shallow
  assert_output --partial "Server does not support shallow clone; performing full clone"
  refute_output --partial "not falling back to full-history clone"
}

# === DoH output IP validation (review finding #8) ===

@test "resolve_alt_ips - skips non-IP (CNAME) answers, returns only valid IPv4" {
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"

  local mock_dir="/tmp/slow_mock_net_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<'SCRIPT'
#!/bin/bash
# First answer is a CNAME hostname (must be ignored), second is a real IP.
echo '{"Answer":[{"data":"github.map.fastly.net."},{"data":"185.199.108.133"}]}'
SCRIPT
  chmod +x "$mock_dir/curl"
  export PATH="$mock_dir:$PATH"

  run checkout::resolve_alt_ips "github.com" "140.82.121.4"
  assert_success
  assert_output "185.199.108.133"
  refute_output --partial "fastly"
}

@test "resolve_alt_ips - returns nothing when no answer is a valid IP" {
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"

  local mock_dir="/tmp/slow_mock_net_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<'SCRIPT'
#!/bin/bash
echo '{"Answer":[{"data":"evil.example.com"},{"data":"999.999.1.1"},{"data":"$(touch /tmp/pwned)"}]}'
SCRIPT
  chmod +x "$mock_dir/curl"
  export PATH="$mock_dir:$PATH"

  run checkout::resolve_alt_ips "github.com" "140.82.121.4"
  assert_success
  assert_output ""
  [ ! -f /tmp/pwned ]
}

@test "is_valid_ipv4 - accepts good addresses, rejects bad ones" {
  run checkout::is_valid_ipv4 "185.199.108.133"
  assert_success
  run checkout::is_valid_ipv4 "0.0.0.0"
  assert_success
  run checkout::is_valid_ipv4 "256.1.1.1"
  assert_failure
  run checkout::is_valid_ipv4 "1.2.3"
  assert_failure
  run checkout::is_valid_ipv4 "github.com"
  assert_failure
  run checkout::is_valid_ipv4 "1.2.3.4; rm -rf /"
  assert_failure
  run checkout::is_valid_ipv4 ""
  assert_failure
}

@test "clone_with_alt_ip - refuses a non-IP alternative endpoint" {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"

  run checkout::clone_with_alt_ip "not-an-ip" "github.com" "443" git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_failure
  assert_output --partial "Refusing to use invalid alternative IP"
}

# === Legacy env-var compatibility (review R2#3) ===

@test "legacy env - SLOW_THRESHOLD/SLOW_TIMEOUT map onto http.lowSpeed* injection" {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=2048
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=25

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "ARGS=$*"
mkdir -p "${@: -1}"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::clone_with_speed_check git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "http.lowSpeedLimit=2048"
  assert_output --partial "http.lowSpeedTime=25"
}

@test "legacy env - new LOW_SPEED_* take precedence over deprecated vars" {
  export SEMAPHORE_GIT_URL="https://github.com/mojombo/grit.git"
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=2048
  export SEMAPHORE_GIT_CLONE_LOW_SPEED_LIMIT=4096

  local mock_dir="/tmp/slow_mock_bin_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/git" <<'SCRIPT'
#!/bin/bash
echo "ARGS=$*"
mkdir -p "${@: -1}"
SCRIPT
  chmod +x "$mock_dir/git"
  export PATH="$mock_dir:$PATH"

  run checkout::clone_with_speed_check git clone "${SEMAPHORE_GIT_URL}" "${SEMAPHORE_GIT_DIR}"
  assert_success
  assert_output --partial "http.lowSpeedLimit=4096"
  refute_output --partial "http.lowSpeedLimit=2048"
}

@test "legacy env - warns once when deprecated SLOW_THRESHOLD/SLOW_TIMEOUT are set" {
  export SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD=2048
  export SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT=25

  run checkout::warn_legacy_env
  assert_output --partial "SEMAPHORE_GIT_CLONE_SLOW_THRESHOLD is deprecated"
  assert_output --partial "SEMAPHORE_GIT_CLONE_SLOW_TIMEOUT is deprecated"
}

@test "legacy env - no deprecation warning when new vars are used" {
  export SEMAPHORE_GIT_CLONE_LOW_SPEED_LIMIT=4096
  export SEMAPHORE_GIT_CLONE_LOW_SPEED_TIME=25

  run checkout::warn_legacy_env
  refute_output --partial "deprecated"
}

# === Stall-kill is opt-in by default (review C1 / R2#2) ===

@test "stall guard - is disabled by default (no STALL_TIMEOUT => no kill)" {
  # Without an explicit opt-in, a long zero-growth phase must be tolerated.
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
dir="$1"
mkdir -p "$dir"
sleep 6
exit 0
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  assert_success
  refute_output --partial "stalled"
}

@test "stall guard - is ON by default for SSH clones (no low-speed gate there)" {
  # SSH has no http.lowSpeedLimit gate, so a hung SSH clone must still be
  # bounded without any explicit opt-in (R2-Med). The default is configurable
  # via SEMAPHORE_GIT_CLONE_SSH_STALL_TIMEOUT (small here to keep the test fast).
  export SEMAPHORE_GIT_URL="git@github.com:mojombo/grit.git"
  export SEMAPHORE_GIT_CLONE_SSH_STALL_TIMEOUT=2
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=0
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1
  # NOTE: SEMAPHORE_GIT_CLONE_STALL_TIMEOUT deliberately NOT set.

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
sleep 120
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  [ "$status" -eq 1 ]
  assert_output --partial "[checkout] Clone stalled"
}

@test "speed check - watchdog does NOT fire while still within the grace window" {
  # Zero growth and a stall guard that WOULD fire post-grace, but the process
  # exits before grace elapses -> must succeed without a kill (grace boundary).
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=2
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=10
  export SEMAPHORE_GIT_CLONE_CHECK_INTERVAL=1

  local mock="/tmp/slow_mock_$$"
  cat > "$mock" <<'SCRIPT'
#!/bin/bash
# Creates nothing (zero growth) and exits well inside the grace window.
sleep 3
exit 0
SCRIPT
  chmod +x "$mock"

  run checkout::clone_with_speed_check "$mock" "$SEMAPHORE_GIT_DIR"
  assert_success
  refute_output --partial "stalled"
}

# === Job-control state is restored (review R2#4) ===

@test "speed check - preserves caller's monitor (set -m) mode when it was on" {
  # Call WITHOUT bats `run` so the function executes in this shell and we can
  # observe $- afterwards (run would mask it in a subshell).
  set -m
  checkout::clone_with_speed_check true >/dev/null 2>&1
  local still_on=0
  case "$-" in *m*) still_on=1 ;; esac
  set +m
  [ "$still_on" -eq 1 ]
}

@test "speed check - does not enable monitor mode when caller had it off" {
  set +m
  checkout::clone_with_speed_check true >/dev/null 2>&1
  local leaked=0
  case "$-" in *m*) leaked=1 ;; esac
  [ "$leaked" -eq 0 ]
}

# === DoH provider fallback (review L5) ===

@test "resolve_alt_ips - falls back to Cloudflare DoH when Google fails" {
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"

  local mock_dir="/tmp/slow_mock_net_$$"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<'SCRIPT'
#!/bin/bash
for a in "$@"; do
  case "$a" in
    *dns.google*)          exit 22 ;;   # Google DoH unreachable (curl -f failure)
    *cloudflare-dns.com*)  echo '{"Answer":[{"data":"185.199.108.133"}]}'; exit 0 ;;
  esac
done
exit 22
SCRIPT
  chmod +x "$mock_dir/curl"
  export PATH="$mock_dir:$PATH"

  run checkout::resolve_alt_ips github.com "140.82.121.4"
  assert_success
  assert_output "185.199.108.133"
}

@test "resolve_alt_ips - sends Accept: application/dns-json header to Cloudflare" {
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="74.0.0.0/8"
  export SEMAPHORE_GIT_CLONE_DOH_PROVIDERS="cloudflare"

  local mock_dir="/tmp/slow_mock_net_$$"
  local args_file="${mock_dir}/curl_args"
  mkdir -p "$mock_dir"
  cat > "$mock_dir/curl" <<SCRIPT
#!/bin/bash
printf '%s\n' "\$@" >> "${args_file}"
echo '{"Answer":[{"data":"185.199.108.133"}]}'
SCRIPT
  chmod +x "$mock_dir/curl"
  export PATH="$mock_dir:$PATH"

  run checkout::resolve_alt_ips github.com "140.82.121.4"
  assert_success
  assert_output "185.199.108.133"

  run cat "${args_file}"
  assert_line "Accept: application/dns-json"
  assert_output --partial "cloudflare-dns.com"
}

# === Observability tokens (review L2) ===

@test "resilient_event - emits a structured greppable token" {
  run checkout::resilient_event tier2_enter
  assert_success
  assert_output "[checkout] event=tier2_enter"
}

@test "resilient_event - appends a metric line when metrics are enabled" {
  export SEMAPHORE_TOOLBOX_METRICS_ENABLED=true
  export SEMAPHORE_GIT_PROVIDER=github
  rm -f /tmp/toolbox_metrics

  checkout::resilient_event tier2_success >/dev/null

  run cat /tmp/toolbox_metrics
  assert_output --partial "libcheckout_resilient"
  assert_output --partial "event=tier2_success"
  rm -f /tmp/toolbox_metrics
}
