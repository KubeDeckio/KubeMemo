function Remove-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [string]$Id,
        [switch]$ExpiredRuntimeOnly,
        [switch]$RemoveAnnotations,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    if ($ExpiredRuntimeOnly) {
        foreach ($note in Get-KubeMemo -IncludeRuntime | Where-Object { $_.StoreType -eq 'Runtime' -and (Test-KubeMemoExpiry -ExpiresAt $_.ExpiresAt) }) {
            if ($PSCmdlet.ShouldProcess($note.Id, 'Delete expired runtime note')) {
                Invoke-KubeMemoKubectl -Arguments @('delete', 'runtimememos.runtime.notes.kubememo.io', $note.Id, '-n', $RuntimeNamespace, '--ignore-not-found=true') | Out-Null
            }
        }
        return
    }

    if (-not $Id) {
        throw 'Specify -Id or use -ExpiredRuntimeOnly.'
    }

    $notes = Get-KubeMemo -IncludeRuntime
    $note = $notes | Where-Object Id -eq $Id | Select-Object -First 1
    if (-not $note) {
        throw "KubeMemo note '$Id' was not found."
    }

    if ($note.StoreType -eq 'Durable' -and -not (Test-KubeMemoDurableWriteAllowed)) {
        throw 'Durable note deletion is blocked in GitOps mode.'
    }

    $resource = if ($note.StoreType -eq 'Runtime') { 'runtimememos.runtime.notes.kubememo.io' } else { 'memos.notes.kubememo.io' }
    $namespace = if ($note.StoreType -eq 'Runtime') { $RuntimeNamespace } else { $note.RawResource.metadata.namespace }
    if ($PSCmdlet.ShouldProcess($Id, 'Delete KubeMemo note')) {
        Invoke-KubeMemoKubectl -Arguments @('delete', $resource, $Id, '-n', $namespace, '--ignore-not-found=true') | Out-Null
        if ($RemoveAnnotations -and $note.TargetMode -eq 'resource') {
            Remove-KubeMemoResourceAnnotations -Kind $note.Kind -Name $note.Name -Namespace $note.Namespace -WhatIf:$WhatIfPreference -Confirm:$false
        }
    }
}
