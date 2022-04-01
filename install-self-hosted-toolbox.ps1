function Install-PSModule {
  param (
    [string] $ModuleLocation,
    [string] $ModuleName
  )

  Write-Output "Installing $ModuleName module in $ModuleLocation..."
  if (-not (Test-Path $ModuleLocation)) {
    Write-Output "No $ModuleLocation directory found. Creating it..."
    New-Item -ItemType Directory -Path $ModuleLocation > $null
    if (-not (Test-Path $ModuleLocation)) {
      Write-Output "Error creating $ModuleLocation"
      Exit 1
    }
  }

  $ModulePath = Join-Path -Path $ModuleLocation -ChildPath $ModuleName
  if (Test-Path $ModulePath) {
    Write-Output "$ModuleName module directory already exists. Overriding it..."
    Remove-Item -Path $ModulePath -Force -Recurse
  }

  Write-Output "Creating $ModuleName module directory at $ModulePath..."
  New-Item -ItemType Directory -Path $ModulePath > $null
  if (-not (Test-Path $ModulePath)) {
    Write-Output "Error creating $ModulePath"
    Exit 1
  }

  Write-Output "Copying .psm1 file to $ModuleName module directory..."

  # The .psm1 file name needs to match the module directory name, otherwise powershell will ignore it
  Copy-Item "$HOME\.toolbox\$ModuleName.psm1" -Destination "$ModulePath\$ModuleName.psm1"
  if (-not $?) {
    Write-Output "Error copying .psm1 module to $ModulePath"
    Exit 1
  }
}

# To make the binaries available,
# we include the $HOME\.toolbox\bin directory in the user's PATH.
function Install-BinaryFolder {
  $toolboxPath = Join-Path $HOME ".toolbox" | Join-Path -ChildPath "bin"
  Write-Output "Adding $toolboxPath to the PATH..."

  $currentPaths = [Environment]::GetEnvironmentVariable('PATH', 'User') -split ';'
  $updatePaths = @($currentPaths | Where-Object { $_ -ne $toolboxPath })
  $updatePaths += $toolboxPath

  [Environment]::SetEnvironmentVariable('PATH', ($updatePaths -join ';'), 'User')
}

$ErrorActionPreference = "Stop"

# PowerShell Core 6.0+ has $isWindows set, but older versions don't.
# For those versions, we use the $env:OS variable, which is only set in Windows.
if ($IsWindows -or $env:OS) {
  $ModulePath = $Env:PSModulePath.Split(";")[0]
} else {
  $ModulePath = $Env:PSModulePath.Split(":")[0]
}

Install-PSModule -ModuleLocation $ModulePath -ModuleName 'Checkout'
Install-BinaryFolder

Write-Output "Installation completed successfully."
