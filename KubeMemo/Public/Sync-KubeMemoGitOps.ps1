function Sync-KubeMemoGitOps {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [string]$Path = (Get-KubeMemoConfigInternal).GitOpsRepoPath,
        [ValidateSet('Export', 'Import')]
        [string]$Direction = 'Export'
    )

    if ($Direction -eq 'Export') {
        Export-KubeMemo -Path $Path
        return
    }

    Import-KubeMemo -Path $Path -WhatIf:$WhatIfPreference -Confirm:$false
}
