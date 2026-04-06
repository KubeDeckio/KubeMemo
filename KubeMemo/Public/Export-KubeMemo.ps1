function Export-KubeMemo {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Path,
        [switch]$IncludeRuntime
    )

    $null = New-Item -Path $Path -ItemType Directory -Force
    foreach ($note in Get-KubeMemo -IncludeRuntime:$IncludeRuntime) {
        $targetPath = Join-Path -Path $Path -ChildPath "$($note.Id).yaml"
        ConvertTo-KubeMemoYaml -Resource ($note.RawResource | ConvertTo-Json -Depth 12 | ConvertFrom-Json -AsHashtable) | Set-Content -Path $targetPath
        Get-Item -Path $targetPath
    }
}
