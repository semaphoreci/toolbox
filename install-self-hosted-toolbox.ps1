$ErrorActionPreference = "Stop"

function Install-PSModules {

  # PowerShell Core 6.0+ has $isWindows set, but older versions don't.
  # For those versions, we use the $env:OS variable, which is only set in Windows.
  if ($IsWindows -or $env:OS) {
    $ModulePath = $Env:PSModulePath.Split(";")[0]
  } else {
    $ModulePath = $Env:PSModulePath.Split(":")[0]
  }

  Write-Output "Installing Checkout module in $ModulePath..."
  if (-not (Test-Path $ModulePath)) {
    Write-Output "No $ModulePath directory found. Creating it..."
    New-Item -ItemType Directory -Path $ModulePath > $null
    if (-not (Test-Path $ModulePath)) {
      Write-Output "Error creating $ModulePath"
      Exit 1
    }
  }

  $CheckoutModulePath = Join-Path -Path $ModulePath -ChildPath "Checkout"
  if (Test-Path $CheckoutModulePath) {
    Write-Output "Checkout module directory already exists. Overriding it..."
    Remove-Item -Path $CheckoutModulePath -Force -Recurse
  }

  Write-Output "Creating Checkout module directory at $CheckoutModulePath..."
  New-Item -ItemType Directory -Path $CheckoutModulePath > $null
  if (-not (Test-Path $CheckoutModulePath)) {
    Write-Output "Error creating $CheckoutModulePath"
    Exit 1
  }

  Write-Output "Copying .psm1 file to checkout module directory..."

  # The .psm1 file name needs to match the module directory name, otherwise powershell will ignore it
  Copy-Item $HOME\.toolbox\Checkout.psm1 -Destination "$CheckoutModulePath\Checkout.psm1"
  if (-not $?) {
    Write-Output "Error copying .psm1 module to $CheckoutModulePath"
    Exit 1
  }
}

# To make the binaries available,
# we include the .toolbox directory in the user's PATH.
function Install-Binaries {
  $toolboxPath = "$HOME\.toolbox"
  Write-Output "Adding $toolboxPath to the PATH..."

  $currentPaths = [Environment]::GetEnvironmentVariable('PATH', 'User') -split ';'
  $updatePaths = @($currentPaths | Where-Object { $_ -ne $toolboxPath })
  $updatePaths += $toolboxPath

  [Environment]::SetEnvironmentVariable('PATH', ($updatePaths -join ';'), 'User')
}

Install-PSModules
Install-Binaries

Write-Output "Installation completed successfully."