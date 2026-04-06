function Test-KubeMemoRbacInstalled {
    [CmdletBinding()]
    param()

    $null = Invoke-KubeMemoKubectl -Arguments @('get', 'clusterrole', 'kubememo-reader', '-o', 'name') -IgnoreErrors
    return ($LASTEXITCODE -eq 0)
}
