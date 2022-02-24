$ErrorActionPreference = "Stop"

$ModulePath = $Env:PSModulePath.Split(";")[0]
Write-Output "Installing libcheckout module in $ModulePath..."
if (-not (Test-Path $ModulePath)) {
  Write-Output "No $ModulePath directory found. Creating it..."
  New-Item -ItemType Directory -Path $ModulePath > $null
  if (-not (Test-Path $ModulePath)) {
    Write-Output "Error creating $ModulePath"
    Exit 1
  }
}

$CheckoutModulePath = $ModulePath + "\Checkout"
if (Test-Path $CheckoutModulePath) {
  Write-Output "libcheckout module directory already exists. Overriding it..."
  Remove-Item -Path $CheckoutModulePath -Force -Recurse
}

Write-Output "Creating libcheckout module directory at $CheckoutModulePath..."
New-Item -ItemType Directory -Path $CheckoutModulePath > $null
if (-not (Test-Path $CheckoutModulePath)) {
  Write-Output "Error creating $CheckoutModulePath"
  Exit 1
}

Write-Output "Copying .psm1 file to checkout module directory..."

# The .psm1 file name needs to match the module directory name, otherwise powershell will ignore it
Copy-Item $HOME\.toolbox\libcheckout.psm1 -Destination "$CheckoutModulePath\Checkout.psm1"
if (-not $?) {
  Write-Output "Error copying .psm1 module to $CheckoutModulePath"
  Exit 1
}

Write-Output "Installation completed successfully."