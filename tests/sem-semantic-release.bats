#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  source sem-semantic-release
  export BATS=true
}

@test "semantic-release::parse_args --help" {
  run semantic-release::parse_args --help

  assert_success
  assert_output --partial "Usage: sem-semantic-release [OPTION]..."
}

@test "semantic-release::parse_args --dry-run" {
  run semantic-release::parse_args --dry-run

  assert_success
  assert_output --partial "semantic-release options: --dry-run "
}

@test "semantic-release::parse_args --version" {
  run semantic-release::parse_args --version 19.0.2

  assert_success
  assert_output --partial "semantic-release version: 19.0.2"
}

@test "semantic-release::parse_args --plugins" {
  run semantic-release::parse_args --plugins @semantic-release/foo @semantic-release-bar

  assert_success
  assert_output --partial "semantic-release plugins: @semantic-release/foo @semantic-release-bar"
}

@test "semantic-release::parse_args --branches" {
  run semantic-release::parse_args --branches master develop release/\*

  assert_success
  assert_output --partial "semantic-release options: --branches master develop release/* "
}

@test "semantic-release::parse_args all options" {
  run semantic-release::parse_args --dry-run --version 19.0.2 --plugins @semantic-release/git --branches master

  assert_success
  assert_output --partial "semantic-release version: 19.0.2"
  assert_output --partial "semantic-release plugins: @semantic-release/git"
  assert_output --partial "semantic-release options: --dry-run --branches master"  
}

@test "semantic-release::install with empty version" {
  run semantic-release::install

  assert_success
  assert [ -e "package.json" ]
  assert [ ! -z $(npx semantic-release --version) ]

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::install with non-empty version" {
  export SEMANTIC_RELEASE_VERSION=19.0.1
  run semantic-release::install

  assert_success 
  assert [ -e "package.json" ]
  assert [ $(npx semantic-release --version) = "19.0.1" ]

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::install with invalid version" {
  export SEMANTIC_RELEASE_VERSION=2122.0.1
  run semantic-release::install

  assert_failure 
  assert_output "sem-semantic-release: Unsupported semantic-release version: 2122.0.1"

  run rm -rf ./node_modules ./package.json ./package-lock.json
}


@test "semantic-release::install with plugins" {
  export SEMANTIC_RELEASE_PLUGINS=("@semantic-release/git@10.0.0" "@semantic-release/changelog")
  run semantic-release::install

  assert_success 
  assert [ -e "package.json" ]

  assert [ $(npm list | grep -oE '@semantic-release/changelog') = "@semantic-release/changelog" ]
  assert [ $(npm list | grep -oE '@semantic-release/git@(.*)') = "@semantic-release/git@10.0.0" ]

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::install with wrong plugins" {
  export SEMANTIC_RELEASE_PLUGINS=("@semantic-release/foo@1.0.0")
  run semantic-release::install

  assert_failure 
  assert_output "sem-semantic-release: Unable to install plugins: @semantic-release/foo@1.0.0"

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::scrape_version with existing version line" {
  echo "The next release version is 2.0.3" > /tmp/semantic-release.log
  run semantic-release::scrape_version

  assert_success
  assert_output "sem-semantic-release: RELEASE_VERSION=2.0.3"
}

@test "semantic-release::scrape_version with non-existing version line" {
  echo "Nothing really happens..." > /tmp/semantic-release.log
  run semantic-release::scrape_version

  assert_success
  assert_output "sem-semantic-release: RELEASE_VERSION not found"
}
