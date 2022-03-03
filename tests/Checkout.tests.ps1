BeforeAll {
  Import-Module Checkout
}

Describe 'Initialize-Repository' {
  BeforeEach {
    $env:SEMAPHORE_GIT_URL = "https://github.com/mojombo/grit.git"
    $env:SEMAPHORE_GIT_BRANCH = "master"
    $env:SEMAPHORE_GIT_DIR = "repo"
    $env:SEMAPHORE_GIT_SHA = "5608567"
    $env:SEMAPHORE_GIT_REF = " refs/heads/master"

    if (Test-Path env:SEMAPHORE_GIT_REF_TYPE) { Remove-Item -Path env:SEMAPHORE_GIT_REF_TYPE }
    if (Test-Path env:SEMAPHORE_GIT_TAG_NAME) { Remove-Item -Path env:SEMAPHORE_GIT_TAG_NAME }
  }

  Context "Required parameters" {
    It 'Fails if SEMAPHORE_GIT_URL is not set' {
      Remove-Item -Path env:SEMAPHORE_GIT_URL
      { Initialize-Repository } | Should -Throw "SEMAPHORE_GIT_URL is required"
    }

    It 'Fails if SEMAPHORE_GIT_BRANCH is not set' {
      Remove-Item -Path env:SEMAPHORE_GIT_BRANCH
      { Initialize-Repository } | Should -Throw "SEMAPHORE_GIT_BRANCH is required"
    }

    It 'Fails if SEMAPHORE_GIT_DIR is not set' {
      Remove-Item -Path env:SEMAPHORE_GIT_DIR
      { Initialize-Repository } | Should -Throw "SEMAPHORE_GIT_DIR is required"
    }

    It 'Fails if SEMAPHORE_GIT_SHA is not set' {
      Remove-Item -Path env:SEMAPHORE_GIT_SHA
      { Initialize-Repository } | Should -Throw "SEMAPHORE_GIT_SHA is required"
    }
  }

  Context "SEMAPHORE_GIT_REF_TYPE = push" {
    It 'branch and SHA exists => success' {
      $env:SEMAPHORE_GIT_REF_TYPE = "push"
      $env:SEMAPHORE_GIT_SHA = "91940c2cc18ec08b751482f806f1b8bfa03d98a5"
      $output = Initialize-Repository
      $output | Should -Contain "Performing shallow clone with depth: 50"
      $output | Should -Contain "HEAD is now at 91940c2 Release 2.4.1"
      $output | Should -Not -Contain "Branch not found, performing full clone"
    }

    It 'branch does not exist => full clone' {
      $env:SEMAPHORE_GIT_REF_TYPE = "push"
      $env:SEMAPHORE_GIT_SHA = "91940c2cc18ec08b751482f806f1b8bfa03d98a5"
      $env:SEMAPHORE_GIT_BRANCH = "this-branch-does-not-exist"
      $output = Initialize-Repository
      $output | Should -Contain "Performing shallow clone with depth: 50"
      $output | Should -Contain "HEAD is now at 91940c2 Release 2.4.1"
      $output | Should -Contain "Branch not found, performing full clone"
    }

    It 'SHA does not exist => error' {
      $env:SEMAPHORE_GIT_REF_TYPE = "push"
      $env:SEMAPHORE_GIT_SHA = "this-sha-does-not-exist"
      { Initialize-Repository } | Should -Throw "SHA: this-sha-does-not-exist not found"
    }
  }

  Context "SEMAPHORE_GIT_REF_TYPE = tag" {
    It 'Tag name and SHA exists => success' {
      $env:SEMAPHORE_GIT_REF_TYPE = "tag"
      $env:SEMAPHORE_GIT_TAG_NAME = "v2.4.1"
      $env:SEMAPHORE_GIT_SHA = "91940c2cc18ec08b751482f806f1b8bfa03d98a5"
      $output = Initialize-Repository
      $output | Should -Contain "Initializing repository for tag..."
      $output | Should -Contain "HEAD is now at 91940c2cc18ec08b751482f806f1b8bfa03d98a5 Release v2.4.1"
    }

    It 'tag does not exist => error' {
      $env:SEMAPHORE_GIT_REF_TYPE = "tag"
      $env:SEMAPHORE_GIT_TAG_NAME = "v9.4.1"
      $env:SEMAPHORE_GIT_SHA = "91940c2cc18ec08b751482f806f1b8bfa03d98a5"
      { Initialize-Repository } | Should -Throw "Release v9.4.1 not found"
    }
  }

  Context "SEMAPHORE_GIT_REF_TYPE = pull-request" {
    It 'SEMAPHORE_GIT_REF exists => success' {
      $env:SEMAPHORE_GIT_REF_TYPE = "pull-request"
      $env:SEMAPHORE_GIT_REF = "refs/pull/186/merge"
      $env:SEMAPHORE_GIT_SHA = "30774365e11f2b1e18706c9ed0920369f6d7c205"
      $output = Initialize-Repository
      $output | Should -Contain "Initializing repository for pull-request..."
      $output | Should -Contain "HEAD is now at 30774365e11f2b1e18706c9ed0920369f6d7c205"
    }

    It 'SEMAPHORE_GIT_REF does not exist => error' {
      $env:SEMAPHORE_GIT_REF_TYPE = "pull-request"
      $env:SEMAPHORE_GIT_REF = "refs/pull/123456789/does-not-exist"
      { Initialize-Repository } | Should -Throw "Revision: $env:SEMAPHORE_GIT_SHA not found"
    }
  }

  Context "No SEMAPHORE_GIT_REF_TYPE" {
    It "old revision => success" {
      $env:SEMAPHORE_GIT_BRANCH = "patch-id"
      $env:SEMAPHORE_GIT_SHA = "da70719"
      $output = Initialize-Repository
      $output | Should -Contain "Performing shallow clone with depth: 50"
      $output | Should -Contain "SHA: da70719 not found, performing full clone"
      $output | Should -Contain "HEAD is now at da70719 Regenerated gemspec for version 2.0.0"
    }

    It "tag => success" {
      $env:SEMAPHORE_GIT_BRANCH = "v2.5.0"
      $env:SEMAPHORE_GIT_SHA = "7219ef6"
      $output = Initialize-Repository
      $output | Should -Contain "Performing shallow clone with depth: 50"
      $output | Should -Contain "HEAD is now at 7219ef6 Release 2.5.0"
    }

    It "refs/tags => success" {
      $env:SEMAPHORE_GIT_BRANCH = "refs/tags/v2.5.0"
      $env:SEMAPHORE_GIT_SHA = "7219ef6"
      $output = Initialize-Repository
      $output | Should -Contain "Performing shallow clone with depth: 50"
      $output | Should -Contain "Branch not found, performing full clone"
      $output | Should -Contain "HEAD is now at 7219ef6 Release 2.5.0"
    }

    It "non-existing SHA => error" {
      $env:SEMAPHORE_GIT_SHA = "1234567"
      { Initialize-Repository } | Should -Throw "SHA: 1234567 not found"
    }
  }
}
