Install-Module -Name PSScriptAnalyzer -Force
if ($?) {
  Invoke-ScriptAnalyzer *
}
