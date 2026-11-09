#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  TEST_HOME=$(mktemp -d)
  sed -n '/^echo "Routing Bitbucket SSH/,/^fi$/p' "$BATS_TEST_DIRNAME/../install-self-hosted-toolbox" > "$TEST_HOME/bitbucket-ssh.sh"
}

teardown() {
  rm -rf "$TEST_HOME"
}

route_bitbucket_ssh() {
  HOME="$TEST_HOME" bash "$TEST_HOME/bitbucket-ssh.sh"
}

@test "install-self-hosted-toolbox: routes bitbucket.org to ssh.bitbucket.org and keeps host key checking" {
  run route_bitbucket_ssh
  assert_success

  run ssh -G -F "$TEST_HOME/.ssh/config" bitbucket.org
  assert_success
  assert_line "hostname ssh.bitbucket.org"
  assert_line "hostkeyalias bitbucket.org"
  refute_line "stricthostkeychecking false"
  refute_line "userknownhostsfile /dev/null"
}

@test "install-self-hosted-toolbox: adds the Bitbucket SSH routing only once" {
  route_bitbucket_ssh
  run route_bitbucket_ssh
  assert_output --partial "Bitbucket SSH routing is already in"

  run grep -c "^Host bitbucket.org$" "$TEST_HOME/.ssh/config"
  assert_output "1"
}

@test "install-self-hosted-toolbox: routes an existing Host bitbucket.org entry and keeps its settings" {
  mkdir -p "$TEST_HOME/.ssh"
  printf 'Host github.com bitbucket.org\n  IdentityFile ~/.ssh/deploy_key\n' > "$TEST_HOME/.ssh/config"

  run route_bitbucket_ssh
  assert_success

  run ssh -G -F "$TEST_HOME/.ssh/config" bitbucket.org
  assert_success
  assert_line "hostname ssh.bitbucket.org"
  assert_line "identityfile ~/.ssh/deploy_key"
}

@test "install-self-hosted-toolbox: keeps a HostName the existing entry sets" {
  mkdir -p "$TEST_HOME/.ssh"
  printf 'Host bitbucket.org\n  HostName 10.0.0.5\n' > "$TEST_HOME/.ssh/config"

  run route_bitbucket_ssh
  assert_success

  run ssh -G -F "$TEST_HOME/.ssh/config" bitbucket.org
  assert_line "hostname 10.0.0.5"
}
