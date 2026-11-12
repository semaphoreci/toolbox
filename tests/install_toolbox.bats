#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

@test "install-toolbox: bitbucket.org SSH connects to ssh.bitbucket.org" {
  run ssh -G -F "$HOME/.ssh/config" bitbucket.org
  assert_success
  assert_line "hostname ssh.bitbucket.org"
  assert_line "stricthostkeychecking false"
  assert_line "userknownhostsfile /dev/null"
}

@test "install-toolbox: ssh.bitbucket.org skips host key checking" {
  run ssh -G -F "$HOME/.ssh/config" ssh.bitbucket.org
  assert_success
  assert_line "hostname ssh.bitbucket.org"
  assert_line "stricthostkeychecking false"
  assert_line "userknownhostsfile /dev/null"
}

@test "install-toolbox: github.com host is unchanged" {
  run ssh -G -F "$HOME/.ssh/config" github.com
  assert_success
  assert_line "hostname github.com"
  assert_line "stricthostkeychecking false"
  assert_line "userknownhostsfile /dev/null"
}
