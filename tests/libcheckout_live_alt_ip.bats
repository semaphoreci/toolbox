#!/usr/bin/env bats
#
# LIVE end-to-end tests that drive the REAL toolbox entry points
# (checkout::resilient_clone and checkout) against a real public repository
# (semaphoreio/semaphore) over the network — not mocks, and not raw git/curl.
#
# Coverage:
#   * checkout::resolve_alt_ips against the real dns.google DoH endpoint,
#   * checkout::resilient_clone cloning the real repo on a healthy network
#     (tier 1 succeeds),
#   * checkout::resilient_clone recovering via tier 2 (a DoH-resolved alternative
#     IP) when the primary endpoint is unreachable (github.com poisoned in
#     /etc/hosts to force the failover),
#   * the full checkout() flow with SEMAPHORE_GIT_CLONE_SLOW_RETRY=true for a
#     branch (push) and a pull request — each routes its clone through
#     checkout::resilient_clone.
#
# Raw git is used ONLY as test scaffolding (to read an expected SHA, or to
# decide whether to skip); every actual clone goes through the toolbox.
#
# Every test skips (does not fail) when the network/DoH/repo — or, for the
# /etc/hosts test, root — is unavailable, so the file is safe in offline CI.
#
# Run in the test image:
#   make libcheckout.docker.test.live

load "support/bats-support/load"
load "support/bats-assert/load"

REPO_HTTPS="https://github.com/semaphoreio/semaphore.git"
REPO_DEFAULT_BRANCH="main"

setup() {
  unset SEMAPHORE_GIT_REF_TYPE
  unset SEMAPHORE_GIT_TAG_NAME
  unset SEMAPHORE_GIT_CLONE_RETRY_COUNT
  unset SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES
  unset SEMAPHORE_GIT_CLONE_SLOW_GRACE
  unset SEMAPHORE_GIT_CLONE_STALL_TIMEOUT
  unset SEMAPHORE_GIT_CLONE_CHECK_INTERVAL

  export SEMAPHORE_GIT_URL="$REPO_HTTPS"
  # Absolute target dir on the container's tmpfs: keeps the (large) live clones
  # off the bind-mounted workspace and immune to checkout()'s internal `cd`.
  export SEMAPHORE_GIT_DIR="$BATS_TEST_TMPDIR/repo"
  export SEMAPHORE_GIT_DEPTH=1
  export SEMAPHORE_GIT_CLONE_SLOW_RETRY=true

  # Spread GeoDNS edns_client_subnet probes across regions that map to
  # github.com's real serving IPs (140.82.x) so a sibling distinct from this
  # runner's current IP is reliably available.
  export SEMAPHORE_GIT_CLONE_ALT_REGIONS="177.0.0.0/8,110.0.0.0/8,13.0.0.0/8,52.0.0.0/8"

  source ~/.toolbox/libcheckout
  cd "$BATS_TEST_TMPDIR" || exit 1
}

teardown() {
  rm -rf "${SEMAPHORE_GIT_DIR:?}" 2>/dev/null || true
  unpoison_github
}

# --- helpers ---------------------------------------------------------------

require_network() {
  curl -sf --connect-timeout 5 --max-time 8 \
    "https://dns.google/resolve?name=github.com&type=A" >/dev/null 2>&1 \
    || skip "dns.google (DoH) unreachable — no network egress"
  git ls-remote --heads "$REPO_HTTPS" "refs/heads/${REPO_DEFAULT_BRANCH}" >/dev/null 2>&1 \
    || skip "github.com / ${REPO_HTTPS} unreachable"
}

# Scaffolding: read a ref's SHA from the live repo (used for assertions / to
# parameterise checkout(), NOT to perform the clone under test).
remote_sha() { git ls-remote "$REPO_HTTPS" "$1" | awk 'NR==1{print $1}'; }

# Scaffolding: confirm at least one DoH-resolved alt IP actually serves the
# repo, so the tier-2 test only runs when recovery is genuinely possible.
has_working_alt_ip() {
  local cur alt
  cur="$(dig +short github.com | head -1)"
  while IFS= read -r alt; do
    [ -z "$alt" ] && continue
    if git -c "http.curloptResolve=github.com:443:${alt}" \
        ls-remote --heads "$REPO_HTTPS" "refs/heads/${REPO_DEFAULT_BRANCH}" \
        >/dev/null 2>&1; then
      return 0
    fi
  done < <(checkout::resolve_alt_ips github.com "$cur")
  return 1
}

# Rewrite /etc/hosts in place (truncate-and-write): it is a bind mount in the
# Docker harness and cannot be renamed (sed -i fails with EBUSY). Works as root
# (Docker image) or as a non-root user with passwordless sudo (Semaphore VM
# agents), and skips cleanly when neither can write it.
_hosts_write() {
  # Reads the full new content on stdin and replaces /etc/hosts.
  if [ -w /etc/hosts ]; then
    cat > /etc/hosts
  elif command -v sudo >/dev/null 2>&1; then
    sudo tee /etc/hosts >/dev/null
  else
    return 1
  fi
}
poison_github() {
  { cat /etc/hosts; printf '127.0.0.1 github.com # libcheckout-live-test\n'; } \
    | _hosts_write 2>/dev/null || true
  # Verify the marker landed; otherwise skip rather than report a confusing
  # tier-1 success (e.g. read-only /etc/hosts with no sudo).
  grep -q '# libcheckout-live-test' /etc/hosts 2>/dev/null \
    || skip "cannot write /etc/hosts to simulate a dead primary endpoint"
}
unpoison_github() {
  if grep -q '# libcheckout-live-test' /etc/hosts 2>/dev/null; then
    grep -v '# libcheckout-live-test' /etc/hosts | _hosts_write 2>/dev/null || true
  fi
}

# --- tests -----------------------------------------------------------------

@test "live: checkout::resolve_alt_ips returns valid, distinct github.com IPs (real DoH)" {
  require_network

  local cur
  cur="$(dig +short github.com | head -1)"

  run checkout::resolve_alt_ips github.com "$cur"
  assert_success
  [ -n "$output" ] || skip "this network's GeoDNS returns no IP distinct from the current one"

  while IFS= read -r ip; do
    [ -z "$ip" ] && continue
    run checkout::is_valid_ipv4 "$ip"
    assert_success
    [ "$ip" != "$cur" ]
  done <<< "$output"
}

@test "live: checkout::resilient_clone clones the real repo on a healthy network (tier 1)" {
  require_network

  local expected
  expected="$(remote_sha "refs/heads/${REPO_DEFAULT_BRANCH}")"
  [ -n "$expected" ]

  run checkout::resilient_clone --depth 1 -b "${REPO_DEFAULT_BRANCH}" --single-branch \
    "$REPO_HTTPS" "$SEMAPHORE_GIT_DIR"
  assert_success
  # Tier 1 served the clone directly — no alternative-endpoint fallback.
  refute_output --partial "Switching to alternative GitHub endpoint"

  [ -d "${SEMAPHORE_GIT_DIR}/.git" ]
  [ "$(git -C "$SEMAPHORE_GIT_DIR" rev-parse HEAD)" = "$expected" ]
}

@test "live: checkout::resilient_clone recovers via tier 2 when the primary endpoint is dead" {
  require_network
  [ "$(id -u)" -eq 0 ] || skip "needs root to poison /etc/hosts"
  has_working_alt_ip || skip "no working alternative IP from this network"

  local expected
  expected="$(remote_sha "refs/heads/${REPO_DEFAULT_BRANCH}")"
  [ -n "$expected" ]

  # Force the primary (system-DNS) path to fail instantly: 127.0.0.1:443 has no
  # listener -> "connection refused". DoH does not consult /etc/hosts, so tier 2
  # still resolves real IPs and routes around the break with http.curloptResolve.
  poison_github

  export SEMAPHORE_GIT_CLONE_RETRY_COUNT=1     # don't re-try the dead primary
  export SEMAPHORE_GIT_CLONE_ALT_IP_RETRIES=4
  # Generous watchdog so the real tier-2 clone isn't killed during GitHub's
  # server-side compression lull (the primary fails via refused-connect, not
  # the watchdog, so this is safe).
  export SEMAPHORE_GIT_CLONE_SLOW_GRACE=20
  export SEMAPHORE_GIT_CLONE_STALL_TIMEOUT=120

  run checkout::resilient_clone --depth 1 -b "${REPO_DEFAULT_BRANCH}" --single-branch \
    "$REPO_HTTPS" "$SEMAPHORE_GIT_DIR"

  unpoison_github

  assert_success
  assert_output --partial "Switching to alternative GitHub endpoint"
  [ -d "${SEMAPHORE_GIT_DIR}/.git" ]
  [ "$(git -C "$SEMAPHORE_GIT_DIR" rev-parse HEAD)" = "$expected" ]
}

@test "live: full checkout() (push) clones a real branch via the resilient path" {
  require_network

  # Discover a real non-default branch + tip (scaffolding only).
  local line branch expected
  line="$(git ls-remote --heads "$REPO_HTTPS" \
    | grep -v "refs/heads/${REPO_DEFAULT_BRANCH}\$" | head -1)"
  [ -n "$line" ] || skip "repo exposes no non-default branch"
  expected="$(printf '%s' "$line" | awk '{print $1}')"
  branch="$(printf '%s' "$line" | awk '{sub(/^refs\/heads\//,"",$2); print $2}')"

  export SEMAPHORE_GIT_REF_TYPE="push"
  export SEMAPHORE_GIT_BRANCH="$branch"
  export SEMAPHORE_GIT_REF="refs/heads/${branch}"
  export SEMAPHORE_GIT_SHA="$expected"

  run checkout
  assert_success
  [ "$(git -C "$SEMAPHORE_GIT_DIR" rev-parse HEAD)" = "$expected" ]
}

@test "live: full checkout() (pull-request) clones a real PR via the resilient path" {
  require_network

  local line expected pr_ref
  line="$(git ls-remote "$REPO_HTTPS" 'refs/pull/*/head' | head -1)"
  [ -n "$line" ] || skip "repo exposes no pull-request refs"
  expected="$(printf '%s' "$line" | awk '{print $1}')"
  pr_ref="$(printf '%s' "$line" | awk '{print $2}')"

  export SEMAPHORE_GIT_REF_TYPE="pull-request"
  export SEMAPHORE_GIT_BRANCH="$REPO_DEFAULT_BRANCH"
  export SEMAPHORE_GIT_REF="$pr_ref"
  export SEMAPHORE_GIT_SHA="$expected"

  run checkout
  assert_success
  assert_output --partial "HEAD is now at"
  [ "$(git -C "$SEMAPHORE_GIT_DIR" rev-parse HEAD)" = "$expected" ]
}
