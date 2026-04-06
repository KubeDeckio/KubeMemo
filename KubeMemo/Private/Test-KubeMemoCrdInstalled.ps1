function Test-KubeMemoCrdInstalled {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [ValidateSet('Durable', 'Runtime')]
        [string]$Store
    )

    $config = Get-KubeMemoConfigInternal
    $name = if ($Store -eq 'Durable') { $config.DurableCrdName } else { $config.RuntimeCrdName }
    $null = Invoke-KubeMemoKubectl -Arguments @('get', 'crd', $name, '-o', 'name') -IgnoreErrors
    return ($LASTEXITCODE -eq 0)
}
