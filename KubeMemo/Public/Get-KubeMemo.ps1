function Get-KubeMemo {
    [CmdletBinding()]
    param(
        [switch]$IncludeRuntime,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [string[]]$Namespace
    )

    $results = New-Object System.Collections.Generic.List[object]
    $durableResource = 'memos.notes.kubememo.io'
    $runtimeResource = 'runtimememos.runtime.notes.kubememo.io'
    $requestedNamespaces = @($Namespace | Where-Object { $_ } | Select-Object -Unique)

    function Add-KubeMemoResults {
        param(
            [Parameter(Mandatory)]
            [pscustomobject]$Status,

            [Parameter(Mandatory)]
            [string]$StoreType
        )

        if ($Status.ExitCode -ne 0 -or [string]::IsNullOrWhiteSpace($Status.Text)) {
            return
        }

        $parsed = $Status.Text | ConvertFrom-Json
        foreach ($item in @($parsed.items)) {
            $results.Add((ConvertTo-KubeMemoObject -Resource $item -StoreType $StoreType))
        }
    }

    if ($requestedNamespaces.Count -gt 0) {
        foreach ($ns in $requestedNamespaces) {
            $durableStatus = Invoke-KubeMemoKubectl -Arguments @('get', $durableResource, '-n', $ns, '-o', 'json') -IgnoreErrors -PassStatus
            Add-KubeMemoResults -Status $durableStatus -StoreType Durable
        }
    }
    else {
        $durableStatus = Invoke-KubeMemoKubectl -Arguments @('get', $durableResource, '--all-namespaces', '-o', 'json') -IgnoreErrors -PassStatus
        if ($durableStatus.ExitCode -eq 0) {
            Add-KubeMemoResults -Status $durableStatus -StoreType Durable
        }
        elseif ($durableStatus.Text -match 'forbidden|cannot list resource') {
            $fallbackNamespace = Get-KubeMemoCurrentNamespace
            $scopedStatus = Invoke-KubeMemoKubectl -Arguments @('get', $durableResource, '-n', $fallbackNamespace, '-o', 'json') -IgnoreErrors -PassStatus
            Add-KubeMemoResults -Status $scopedStatus -StoreType Durable
        }
    }

    if ($IncludeRuntime) {
        $runtimeStatus = Invoke-KubeMemoKubectl -Arguments @('get', $runtimeResource, '-n', $RuntimeNamespace, '-o', 'json') -IgnoreErrors -PassStatus
        Add-KubeMemoResults -Status $runtimeStatus -StoreType Runtime
    }

    $results
}
