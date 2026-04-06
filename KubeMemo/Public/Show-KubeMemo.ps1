function Show-KubeMemo {
    [CmdletBinding()]
    param(
        [string]$Kind,
        [string]$Namespace,
        [string]$Name,
        [switch]$IncludeRuntime,
        [switch]$PassThru,
        [switch]$NoColor
    )

    $notes = Get-KubeMemo -IncludeRuntime:$IncludeRuntime |
        Get-KubeMemoMatches -Kind $Kind -Namespace $Namespace -Name $Name

    if (-not $notes) {
        return
    }

    Write-KubeMemoCard -Notes $notes -PassThru:$PassThru -NoColor:$NoColor
}
