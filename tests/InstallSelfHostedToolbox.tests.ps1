Describe 'install-self-hosted-toolbox.ps1 Bitbucket SSH routing' {
  BeforeAll {
    $SSHConfigPath = Join-Path (Join-Path $HOME ".ssh") "config"
    $Installer = Join-Path (Join-Path $HOME ".toolbox") "install-self-hosted-toolbox.ps1"
  }

  It 'routes bitbucket.org to ssh.bitbucket.org with the bitbucket.org host keys' {
    $config = Get-Content -Path $SSHConfigPath
    $config | Should -Contain "Host bitbucket.org"
    $config | Should -Contain "  HostName ssh.bitbucket.org"
    $config | Should -Contain "  HostKeyAlias bitbucket.org"
    $config | Should -Not -Contain "  StrictHostKeyChecking no"
  }

  It 'adds the routing only once when installed again' {
    & $Installer | Should -Contain "Bitbucket SSH routing is already in $SSHConfigPath"
    @(Get-Content -Path $SSHConfigPath | Where-Object { $_ -eq "Host bitbucket.org" }).Count | Should -Be 1
  }

  It 'appends the routing after an existing Host bitbucket.org entry' {
    $original = Get-Content -Path $SSHConfigPath -Raw
    try {
      Set-Content -Path $SSHConfigPath -Encoding ascii -Value @("Host github.com bitbucket.org", "  IdentityFile ~/.ssh/deploy_key")
      & $Installer > $null
      $config = @(Get-Content -Path $SSHConfigPath)
      $config[0..1] | Should -Be @("Host github.com bitbucket.org", "  IdentityFile ~/.ssh/deploy_key")
      $config | Should -Contain "  HostName ssh.bitbucket.org"
      $config.IndexOf("Host bitbucket.org") | Should -BeGreaterThan 1
    } finally {
      Set-Content -Path $SSHConfigPath -Encoding ascii -Value $original -NoNewline
    }
  }
}
