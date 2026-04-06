function Clear-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [switch]$ExpiredOnly,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    if ($ExpiredOnly) {
        Remove-KubeMemo -ExpiredRuntimeOnly -RuntimeNamespace $RuntimeNamespace -WhatIf:$WhatIfPreference -Confirm:$false
        return
    }

    if ($PSCmdlet.ShouldProcess($RuntimeNamespace, 'Delete all runtime KubeMemo notes')) {
        Invoke-KubeMemoKubectl -Arguments @('delete', 'runtimememos.runtime.notes.kubememo.io', '--all', '-n', $RuntimeNamespace, '--ignore-not-found=true') | Out-Null
    }
}
