function Checkout {
  $ErrorActionPreference = "Stop"

  Get-Command git > $null
  if (-not $?) {
    Write-Output '[ERROR] Could not find git.exe in $env:Path.'
    Exit 1
  }

  if ($null -eq $env:SEMAPHORE_GIT_BRANCH) {
    Write-Output "[ERROR] SEMAPHORE_GIT_BRANCH is required."
    Exit 1
  }

  if ($null -eq $env:SEMAPHORE_GIT_URL) {
    Write-Output "[ERROR] SEMAPHORE_GIT_URL is required. Exiting..."
    Exit 1
  }

  if ($null -eq $env:SEMAPHORE_GIT_DIR) {
    Write-Output "[ERROR] SEMAPHORE_GIT_DIR is required. Exiting..."
    Exit 1
  }

  if ($null -eq $env:SEMAPHORE_GIT_SHA) {
    Write-Output "[ERROR] SEMAPHORE_GIT_SHA is required. Exiting..."
    Exit 1
  }

  if (Test-Path $env:SEMAPHORE_GIT_DIR) {
    Remove-Item $env:SEMAPHORE_GIT_DIR -Recurse -force
  }

  if (Test-Path env:SEMAPHORE_GIT_REF_TYPE) {
    Ref-Based-Checkout
  } else {
    Shallow-Checkout
  }
}

function Check-Revision {
  git rev-list HEAD..$env:SEMAPHORE_GIT_SHA 2> $null
  if (-not $?) {
    Write-Output "Revision: $env:SEMAPHORE_GIT_SHA} not found .... Exiting"
    return 1
  }
}

function Shallow-Checkout() {
  if ($null -eq $env:SEMAPHORE_GIT_DEPTH) {
    $env:SEMAPHORE_GIT_DEPTH = 50
  }

  Write-Output "Performing shallow clone with depth: $env:SEMAPHORE_GIT_DEPTH"
  git clone --depth $env:SEMAPHORE_GIT_DEPTH -b $env:SEMAPHORE_GIT_BRANCH $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR

  if (-not $?) {
    Write-Output "Branch not found performing full clone"
    git clone $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR
    cd $env:SEMAPHORE_GIT_DIR
    Check-Revision
    if ($?) {
      git reset --hard $env:SEMAPHORE_GIT_SHA 2> $null
    } else {
      return 1
    }
  } else {
    cd $env:SEMAPHORE_GIT_DIR
    git reset --hard $env:SEMAPHORE_GIT_SHA 2> $null
    if (-not $?) {
      Write-Output "SHA: $env:SEMAPHORE_GIT_SHA not found performing full clone"
      git fetch --unshallow
      Check-Revision
      if ($?) {
        git reset --hard $env:SEMAPHORE_GIT_SHA 2> $null
      } else {
        return 1
      }
    }
  }
}

function Ref-Based-Checkout {
  if ($null -eq $env:SEMAPHORE_GIT_DEPTH) {
    $env:SEMAPHORE_GIT_DEPTH = 50
  }

  if ($env:SEMAPHORE_GIT_REF_TYPE -eq "pull-request") {
    git clone --depth $env:SEMAPHORE_GIT_DEPTH $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR 2> $null
    cd $env:SEMAPHORE_GIT_DIR
    git fetch origin +$env:SEMAPHORE_GIT_REF: 2> $null
    if (-not $?) {
      Write-Output "Revision: $env:SEMAPHORE_GIT_SHA not found .... Exiting"
      return 1
    } else {
      git checkout -qf FETCH_HEAD
      Write-Output "HEAD is now at $env:SEMAPHORE_GIT_SHA"
      return 0
    }
  }

  if ($env:SEMAPHORE_GIT_REF_TYPE -eq "tag") {
    git clone --depth $env:SEMAPHORE_GIT_DEPTH -b $env:SEMAPHORE_GIT_TAG_NAME $env:SEMAPHORE_GIT_URL $env:SEMAPHORE_GIT_DIR 2> $null
    if (-not $?) {
      Write-Output "Release $env:SEMAPHORE_GIT_TAG_NAME not found .... Exiting"
      return 1
    } else {
      cd $env:SEMAPHORE_GIT_DIR
      git checkout -qf $env:SEMAPHORE_GIT_TAG_NAME
      Write-Output "HEAD is now at $env:SEMAPHORE_GIT_SHA Release $env:SEMAPHORE_GIT_TAG_NAME"
      return 0
    }
  }

  Shallow-Checkout
}

Export-ModuleMember -Function Checkout -Alias checkout
