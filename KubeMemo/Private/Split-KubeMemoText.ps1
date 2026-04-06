function Split-KubeMemoText {
    [CmdletBinding()]
    param(
        [AllowNull()]
        [string]$Text,

        [int]$Width = 72
    )

    if ([string]::IsNullOrWhiteSpace($Text)) {
        return @('')
    }

    $wrapped = New-Object System.Collections.Generic.List[string]
    foreach ($rawLine in ($Text -split "`r?`n")) {
        $line = $rawLine.TrimEnd()
        if ($line.Length -le $Width) {
            $wrapped.Add($line)
            continue
        }

        $remaining = $line
        while ($remaining.Length -gt $Width) {
            $slice = $remaining.Substring(0, $Width)
            $breakIndex = $slice.LastIndexOf(' ')
            if ($breakIndex -lt [Math]::Floor($Width / 3)) {
                $breakIndex = $Width
            }

            $wrapped.Add($remaining.Substring(0, $breakIndex).TrimEnd())
            $remaining = $remaining.Substring($breakIndex).TrimStart()
        }

        if ($remaining.Length -gt 0) {
            $wrapped.Add($remaining)
        }
    }

    return $wrapped
}
