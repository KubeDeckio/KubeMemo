function Invoke-KubeMemoBinary {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string[]]$Arguments,

        [switch]$ParseJson,

        [switch]$PassthruTerminal
    )

    $binary = Get-KubeMemoBinaryPath
    if ($PassthruTerminal) {
        try {
            & $binary @Arguments
            if ($LASTEXITCODE -ne 0) {
                throw "kubememo $($Arguments -join ' ') failed with exit code $LASTEXITCODE."
            }
        } catch {
            $message = if ($_.Exception -and $_.Exception.Message) { $_.Exception.Message } else { $_ | Out-String }
            throw "Failed to execute kubememo binary at '$binary': $message"
        }
        return
    }

    try {
        $output = & $binary @Arguments 2>&1
    } catch {
        $message = if ($_.Exception -and $_.Exception.Message) { $_.Exception.Message } else { $_ | Out-String }
        throw "Failed to execute kubememo binary at '$binary': $message"
    }
    if ($LASTEXITCODE -ne 0) {
        throw ($output -join [Environment]::NewLine)
    }

    $text = ($output -join [Environment]::NewLine).Trim()
    if (-not $ParseJson) {
        return $text
    }

    if ([string]::IsNullOrWhiteSpace($text)) {
        return $null
    }

    $parsed = $text | ConvertFrom-Json -Depth 20
    if ($parsed.PSObject.Properties.Name -contains 'items') {
        return @($parsed.items)
    }

    return $parsed
}
