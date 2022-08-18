#!/usr/bin/env bats

load "support/bats-support/load"
load "support/bats-assert/load"

setup() {
  export SEMANTIC_RELEASE_PLUGINS=""
  export SEMANTIC_RELEASE_OPTIONS=""
  export SEMANTIC_RELEASE_VERSION=""
}

@test "semantic-release::parse_args --help" {
  source ~/.toolbox/sem-semantic-release
  run semantic-release::parse_args --help

  assert_success
  assert_output --partial "Usage: sem-semantic-release [OPTION]..."
}

@test "semantic-release::parse_args --dry-run" {
  source ~/.toolbox/sem-semantic-release
  run semantic-release::parse_args --dry-run

  assert_success
  assert_output --partial "semantic-release options: --dry-run "
}

@test "semantic-release::parse_args --version" {
  source ~/.toolbox/sem-semantic-release
  run semantic-release::parse_args --version 19.0.2

  assert_success
  assert_output --partial "semantic-release version: 19.0.2"
}

@test "semantic-release::parse_args --plugins" {
  source ~/.toolbox/sem-semantic-release
  run semantic-release::parse_args --plugins @semantic-release/foo @semantic-release-bar

  assert_success
  assert_output --partial "semantic-release plugins: @semantic-release/foo @semantic-release-bar"
}

@test "semantic-release::parse_args --branches" {
  source ~/.toolbox/sem-semantic-release
  run semantic-release::parse_args --branches master develop release/\*

  assert_success
  assert_output --partial "semantic-release options: --branches master develop release/* "
}

@test "semantic-release::parse_args all options" {
  source ~/.toolbox/sem-semantic-release
  run semantic-release::parse_args --dry-run --version 19.0.2 --plugins @semantic-release/git --branches master

  assert_success
  assert_output --partial "semantic-release version: 19.0.2"
  assert_output --partial "semantic-release plugins: @semantic-release/git"
  assert_output --partial "semantic-release options: --dry-run --branches master"  
}

@test "semantic-release::install with empty version" {
  source ~/.toolbox/sem-semantic-release
  export SEMANTIC_RELEASE_PLUGINS=""
  run semantic-release::install

  assert_success
  assert [ -e "package.json" ]
  assert [ ! -z $(npx semantic-release --version) ]

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::install with non-empty version" {
  source ~/.toolbox/sem-semantic-release
  export SEMANTIC_RELEASE_PLUGINS=""
  export SEMANTIC_RELEASE_VERSION=19.0.1
  run semantic-release::install

  assert_success 
  assert [ -e "package.json" ]
  assert [ $(npx semantic-release --version) = "19.0.1" ]

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::install with invalid version" {
  source ~/.toolbox/sem-semantic-release
  export SEMANTIC_RELEASE_PLUGINS=""
  export SEMANTIC_RELEASE_VERSION=2122.0.1
  run semantic-release::install

  assert_failure 
  assert_output "sem-semantic-release: Unsupported semantic-release version: 2122.0.1"

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::install with plugins" {
  source ~/.toolbox/sem-semantic-release
  export SEMANTIC_RELEASE_PLUGINS="@semantic-release/git@10.0.1 @semantic-release/changelog"
  run semantic-release::install

  assert_success 
  assert [ -e "package.json" ]

  assert [ -n $(npm view @semantic-release/changelog version) ]
  assert [ $(npm view @semantic-release/git version) = "10.0.1" ]

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::install with wrong plugins" {
  source ~/.toolbox/sem-semantic-release
  export SEMANTIC_RELEASE_PLUGINS="@semantic-release/foo@1.0.0"
  run semantic-release::install

  assert_failure 
  assert_output "sem-semantic-release: Unable to install plugins: @semantic-release/foo@1.0.0"

  run rm -rf ./node_modules ./package.json ./package-lock.json
}

@test "semantic-release::scrape_version with existing version line" {
  source ~/.toolbox/sem-semantic-release
  echo "The next release version is 2.0.3" > /tmp/semantic-release.log
  export SEMANTIC_RELEASE_RESULT=0
  run semantic-release::scrape_version

  assert_success
  assert_output --partial "Release 2.0.3 has been generated"
}

@test "semantic-release::scrape_version with sem-context get" {
  source ~/.toolbox/sem-semantic-release
  echo "The next release version is 2.0.3" > /tmp/semantic-release.log
  export SEMANTIC_RELEASE_RESULT=0
  run semantic-release::scrape_version

  assert_success
  assert [ $(sem-context get ReleasePublished) = "true" ]
  assert [ $(sem-context get ReleaseVersion) = "2.0.3" ]
  assert [ $(sem-context get ReleaseMajorVersion) = "2" ]
  assert [ $(sem-context get ReleaseMinorVersion) = "0" ]
  assert [ $(sem-context get ReleasePatchVersion) = "3" ]
}

@test "semantic-release::scrape_version with non-existing version line" {
  source ~/.toolbox/sem-semantic-release
  echo "Nothing really happens..." > /tmp/semantic-release.log
  export SEMANTIC_RELEASE_RESULT=0
  run semantic-release::scrape_version

  assert_success
  assert_output --partial "New release hasn't been generated"
}
