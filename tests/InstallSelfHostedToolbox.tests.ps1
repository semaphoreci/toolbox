Describe 'install-self-hosted-toolbox.ps1 Bitbucket SSH routing' {
  BeforeAll {
    $script:SSHConfigPath = Join-Path (Join-Path $HOME ".ssh") "config"
    $script:Installer = Join-Path (Join-Path $HOME ".toolbox") "install-self-hosted-toolbox.ps1"
  }

  It 'routes bitbucket.org to ssh.bitbucket.org with the bitbucket.org host keys' {
    $config = Get-Content -Path $script:SSHConfigPath
    $config | Should -Contain "Host bitbucket.org"
    $config | Should -Contain "  HostName ssh.bitbucket.org"
    $config | Should -Contain "  HostKeyAlias bitbucket.org"
    $config | Should -Not -Contain "  StrictHostKeyChecking no"
  }

  It 'adds the routing only once when installed again' {
    & $script:Installer | Should -Contain "Bitbucket SSH routing is already in $script:SSHConfigPath"
    @(Get-Content -Path $script:SSHConfigPath | Where-Object { $_ -eq "Host bitbucket.org" }).Count | Should -Be 1
  }

  It 'appends the routing after an existing Host bitbucket.org entry' {
    $original = Get-Content -Path $script:SSHConfigPath -Raw
    try {
      Set-Content -Path $script:SSHConfigPath -Encoding ascii -Value @("Host github.com bitbucket.org", "  IdentityFile ~/.ssh/deploy_key")
      & $script:Installer > $null
      $config = @(Get-Content -Path $script:SSHConfigPath)
      $config[0..1] | Should -Be @("Host github.com bitbucket.org", "  IdentityFile ~/.ssh/deploy_key")
      $config | Should -Contain "  HostName ssh.bitbucket.org"
      $config.IndexOf("Host bitbucket.org") | Should -BeGreaterThan 1
    } finally {
      Set-Content -Path $script:SSHConfigPath -Encoding ascii -Value $original -NoNewline
    }
  }
}
