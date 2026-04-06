function Get-KubeMemoCurrentNamespace {
    [CmdletBinding()]
    param()

    $status = Invoke-KubeMemoKubectl -Arguments @('config', 'view', '--minify', '-o', 'jsonpath={..namespace}') -IgnoreErrors -PassStatus
    if ($status.ExitCode -ne 0) {
        return 'default'
    }

    $namespace = $status.Text.Trim()
    if ([string]::IsNullOrWhiteSpace($namespace)) {
        return 'default'
    }

    return $namespace
}
