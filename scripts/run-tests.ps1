$ProgressPreference = 'SilentlyContinue'
$ErrorActionPreference = 'Stop'

$ToolboxPath = "$HOME\.toolbox"
if (Test-Path $ToolboxPath) {
  Remove-Item -Path $ToolboxPath -Force -Recurse
}

New-Item -Path $ToolboxPath -ItemType Directory > $null

# Copy toolbox to right place and install it
Copy-Item -Path Checkout.psm1 -Destination "$ToolboxPath\Checkout.psm1"
Copy-Item -Path .\install-self-hosted-toolbox.ps1 -Destination "$ToolboxPath\install-self-hosted-toolbox.ps1"
& "$ToolboxPath\install-self-hosted-toolbox.ps1"

# Import and run tests using Pester
Install-Module -Name Pester -Force
if ($?) {
  Invoke-Pester -Output Detailed .\tests\Checkout.Tests.ps1
}
