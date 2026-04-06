function Test-KubeMemoNamespaceInstalled {
    [CmdletBinding()]
    param(
        [string]$Namespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    $null = Invoke-KubeMemoKubectl -Arguments @('get', 'namespace', $Namespace, '-o', 'name') -IgnoreErrors
    return ($LASTEXITCODE -eq 0)
}
