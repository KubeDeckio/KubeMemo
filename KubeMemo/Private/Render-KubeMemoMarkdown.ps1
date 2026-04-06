function Render-KubeMemoMarkdown {
    [CmdletBinding()]
    param(
        [AllowNull()]
        [string]$Content
    )

    if (-not $Content) {
        return ''
    }

    $rendered = $Content -replace '^#+\s*', '' -replace '\*\*', '' -replace '`', ''
    return ($rendered -split "`r?`n" | ForEach-Object { $_.TrimEnd() }) -join [Environment]::NewLine
}
