function Uninstall-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [switch]$RuntimeOnly,
        [switch]$RemoveData,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    if ($RuntimeOnly) {
        if ($PSCmdlet.ShouldProcess($RuntimeNamespace, 'Remove KubeMemo runtime namespace')) {
            Invoke-KubeMemoKubectl -Arguments @('delete', 'namespace', $RuntimeNamespace, '--ignore-not-found=true') | Out-Null
        }
        return
    }

    if ($RemoveData) {
        foreach ($resource in @(
            'runtimememos.runtime.notes.kubememo.io',
            'memos.notes.kubememo.io'
        )) {
            if ($PSCmdlet.ShouldProcess($resource, 'Delete all note objects')) {
                Invoke-KubeMemoKubectl -Arguments @('delete', $resource, '--all', '--all-namespaces', '--ignore-not-found=true') | Out-Null
            }
        }
    }

    foreach ($resource in @(
        'crd/runtimememos.runtime.notes.kubememo.io',
        'crd/memos.notes.kubememo.io'
    )) {
        if ($PSCmdlet.ShouldProcess($resource, 'Delete CRD')) {
            Invoke-KubeMemoKubectl -Arguments @('delete', $resource, '--ignore-not-found=true') | Out-Null
        }
    }
}
