# Don't display progress bar when installing PSScriptAnalyzer
$ProgressPreference = 'SilentlyContinue'

Install-Module -Name PSScriptAnalyzer -Force
if ($?) {
  Invoke-ScriptAnalyzer * -EnableExit -ReportSummary
}
