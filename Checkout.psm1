function Initialize-Repository {
  <#
    .Description
    The Initialize-Repository function clones and sets up your Git repository. It requires a few environment variables to work:
      - SEMAPHORE_GIT_BRANCH
      - SEMAPHORE_GIT_URL
      - SEMAPHORE_GIT_DIR
      - SEMAPHORE_GIT_SHA
    if SEMAPHORE_GIT_REF_TYPE is set, a ref-based checkout will be done. If not, a shallow checkout is done.
  #>

  Get-Command git > $null

  if (-not (Test-Path env:SEMAPHORE_GIT_BRANCH)) {
    throw "SEMAPHORE_GIT_BRANCH is required"
  }

  if (-not (Test-Path env:SEMAPHORE_GIT_URL)) {
    throw "SEMAPHORE_GIT_URL is required"
  }

  if (-not (Test-Path env:SEMAPHORE_GIT_DIR)) {
    throw "SEMAPHORE_GIT_DIR is required"
  }

  if (-not (Test-Path env:SEMAPHORE_GIT_SHA)) {
    throw "SEMAPHORE_GIT_SHA is required"
  }

  if (Test-Path $env:SEMAPHORE_GIT_DIR) {
    Remove-Item $env:SEMAPHORE_GIT_DIR -Recurse -force
  }

  if (-not (Test-Path env:SEMAPHORE_GIT_DEPTH)) {
    $env:SEMAPHORE_GIT_DEPTH = 50
  }

  switch ($env:SEMAPHORE_GIT_REF_TYPE) {

    "pull-request" {
      Write-Output "Initializing repository for pull-request..."
      git clone --depth $env:SEMAPHORE_GIT_DEPTH $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR 2> $null
      Set-Location $env:SEMAPHORE_GIT_DIR
      git fetch origin +${env:SEMAPHORE_GIT_REF}: 2> $null
      if (-not $?) {
        throw "Revision: $env:SEMAPHORE_GIT_SHA not found"
      } else {
        git checkout -qf FETCH_HEAD
        Write-Output "HEAD is now at $env:SEMAPHORE_GIT_SHA"
      }
    }

    "tag" {
      Write-Output "Initializing repository for tag..."
      git clone --depth $env:SEMAPHORE_GIT_DEPTH -b $env:SEMAPHORE_GIT_TAG_NAME $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR 2> $null
      if (-not $?) {
        throw "Release $env:SEMAPHORE_GIT_TAG_NAME not found"
      } else {
        Set-Location $env:SEMAPHORE_GIT_DIR
        git checkout -qf $env:SEMAPHORE_GIT_TAG_NAME
        Write-Output "HEAD is now at $env:SEMAPHORE_GIT_SHA Release $env:SEMAPHORE_GIT_TAG_NAME"
      }
    }

    Default {
      Initialize-ShallowRepository
    }
  }
}

function Initialize-ShallowRepository() {
  Write-Output "Performing shallow clone with depth: $env:SEMAPHORE_GIT_DEPTH"
  git clone --depth $env:SEMAPHORE_GIT_DEPTH -b $env:SEMAPHORE_GIT_BRANCH $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR

  if (-not $?) {
    Write-Output "Branch not found, performing full clone"
    git clone $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR
    Set-Location $env:SEMAPHORE_GIT_DIR
    if (Test-Revision) {
      git reset --hard $env:SEMAPHORE_GIT_SHA 2> $null
    } else {
      throw "SHA: $env:SEMAPHORE_GIT_SHA not found"
    }
  } else {
    Set-Location $env:SEMAPHORE_GIT_DIR
    git reset --hard $env:SEMAPHORE_GIT_SHA 2> $null
    if (-not $?) {
      Write-Output "SHA: $env:SEMAPHORE_GIT_SHA not found, performing full clone"
      git fetch --unshallow
      if (Test-Revision) {
        git reset --hard $env:SEMAPHORE_GIT_SHA 2> $null
      } else {
        throw "SHA: $env:SEMAPHORE_GIT_SHA not found"
      }
    }
  }
}

function Test-Revision {
  git rev-list HEAD..$env:SEMAPHORE_GIT_SHA 2> $null
  if (-not $?) {
    return $false
  }

  return $true
}

Export-ModuleMember -Function Initialize-Repository -Alias checkout
