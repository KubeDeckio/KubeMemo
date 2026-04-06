function Get-KubeMemoActivity {
    [CmdletBinding()]
    param(
        [string]$Kind,
        [string]$Namespace,
        [string]$Name
    )

    Get-KubeMemo -IncludeRuntime |
        Where-Object { $_.StoreType -eq 'Runtime' -and $_.NoteType -eq 'activity' } |
        Get-KubeMemoMatches -Kind $Kind -Namespace $Namespace -Name $Name
}
