function Get-KubeMemoMatches {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory, ValueFromPipeline)]
        [psobject[]]$InputObject,

        [string]$Text,
        [string]$NoteType,
        [string]$Kind,
        [string]$Namespace,
        [string]$Name,
        [switch]$IncludeExpired
    )

    process {
        foreach ($item in $InputObject) {
            if (-not $IncludeExpired -and (Test-KubeMemoExpiry -ExpiresAt $item.ExpiresAt)) {
                continue
            }

            if ($Text -and (($item.Title, $item.Summary, $item.Content -join "`n") -notmatch [regex]::Escape($Text))) {
                continue
            }

            if ($NoteType -and $item.NoteType -ne $NoteType) {
                continue
            }

            if ($Kind -and $item.Kind -ne $Kind) {
                continue
            }

            if ($Namespace -and $item.Namespace -ne $Namespace) {
                continue
            }

            if ($Name -and $item.Name -ne $Name) {
                continue
            }

            $item
        }
    }
}
