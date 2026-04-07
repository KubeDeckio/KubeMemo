$privatePath = Join-Path -Path $PSScriptRoot -ChildPath 'Private'
$wrapperPath = Join-Path -Path $PSScriptRoot -ChildPath 'Wrapper/PublicWrapper.ps1'

foreach ($scriptName in 'Get-KubeMemoConfigInternal.ps1', 'Get-KubeMemoBinaryPath.ps1', 'Invoke-KubeMemoBinary.ps1') {
    . (Join-Path -Path $privatePath -ChildPath $scriptName)
}

. $wrapperPath
