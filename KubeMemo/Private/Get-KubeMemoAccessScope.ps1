function Get-KubeMemoAccessScope {
    [CmdletBinding()]
    param(
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    $currentNamespace = Get-KubeMemoCurrentNamespace
    $durable = Invoke-KubeMemoKubectl -Arguments @('get', 'memos.notes.kubememo.io', '--all-namespaces', '-o', 'name') -IgnoreErrors -PassStatus
    $runtime = Invoke-KubeMemoKubectl -Arguments @('get', 'runtimememos.runtime.notes.kubememo.io', '-n', $RuntimeNamespace, '-o', 'name') -IgnoreErrors -PassStatus

    $durableScope = 'cluster'
    if ($durable.ExitCode -ne 0) {
        if ($durable.Text -match 'forbidden|cannot list resource') {
            $durableScope = "namespace:$currentNamespace"
        }
        else {
            $durableScope = 'unavailable'
        }
    }

    [pscustomobject]@{
        CurrentNamespace = $currentNamespace
        DurableScope = $durableScope
        RuntimeReadable = ($runtime.ExitCode -eq 0)
        RuntimeNamespace = $RuntimeNamespace
    }
}
