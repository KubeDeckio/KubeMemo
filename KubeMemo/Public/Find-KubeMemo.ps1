function Find-KubeMemo {
    [CmdletBinding()]
    param(
        [string]$Text,
        [string]$NoteType,
        [string]$Kind,
        [string]$Namespace,
        [string]$Name,
        [switch]$IncludeRuntime,
        [switch]$IncludeExpired
    )

    Get-KubeMemo -IncludeRuntime:$IncludeRuntime |
        Get-KubeMemoMatches -Text $Text -NoteType $NoteType -Kind $Kind -Namespace $Namespace -Name $Name -IncludeExpired:$IncludeExpired
}
