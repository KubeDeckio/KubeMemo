function Test-KubeMemoExpiry {
    [CmdletBinding()]
    param(
        [AllowNull()]
        $ExpiresAt
    )

    if (-not $ExpiresAt) {
        return $false
    }

    $expiryValue = if ($ExpiresAt -is [datetime]) {
        $ExpiresAt
    }
    else {
        try {
            [datetime]::Parse($ExpiresAt.ToString())
        }
        catch {
            return $false
        }
    }

    return [datetime]::UtcNow -ge $expiryValue.ToUniversalTime()
}
