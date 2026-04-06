function Format-KubeMemoRelativeTime {
    [CmdletBinding()]
    param(
        [AllowNull()]
        $Timestamp
    )

    if (-not $Timestamp) {
        return $null
    }

    $value = if ($Timestamp -is [datetime]) {
        $Timestamp
    }
    else {
        try {
            [datetime]::Parse($Timestamp.ToString())
        }
        catch {
            return $Timestamp.ToString()
        }
    }

    $span = $value.ToUniversalTime() - [datetime]::UtcNow
    $prefix = if ($span.TotalSeconds -lt 0) { 'ago' } else { 'in' }
    $abs = [timespan]::FromSeconds([math]::Abs($span.TotalSeconds))

    if ($abs.TotalDays -ge 1) {
        $count = [math]::Floor($abs.TotalDays)
        if ($prefix -eq 'ago') {
            return "$count" + 'd ago'
        }
        return 'in ' + "$count" + 'd'
    }
    if ($abs.TotalHours -ge 1) {
        $count = [math]::Floor($abs.TotalHours)
        if ($prefix -eq 'ago') {
            return "$count" + 'h ago'
        }
        return 'in ' + "$count" + 'h'
    }
    $count = [math]::Max(1, [math]::Floor($abs.TotalMinutes))
    if ($prefix -eq 'ago') {
        return "$count" + 'm ago'
    }
    return 'in ' + "$count" + 'm'
}
