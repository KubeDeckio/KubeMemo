function ConvertTo-KubeMemoYaml {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [hashtable]$Resource
    )

    $json = $Resource | ConvertTo-Json -Depth 12
    $indentWidth = 2

    function Convert-ValueToYaml {
        param(
            [Parameter(Mandatory)]
            $Value,
            [int]$Indent = 0
        )

        $prefix = ' ' * $Indent

        if ($null -eq $Value) {
            return @("${prefix}null")
        }

        if ($Value -is [string]) {
            if ($Value -match "`r?`n") {
                $lines = @("${prefix}|")
                foreach ($line in ($Value -split "`r?`n")) {
                    $lines += (' ' * ($Indent + $indentWidth)) + $line
                }
                return $lines
            }

            if ($Value -eq '' -or $Value -match '[:#\-\{\}\[\],&\*!\|>%@`"]' -or $Value.Trim() -ne $Value) {
                $escaped = $Value.Replace('"', '\"')
                return @("${prefix}`"$escaped`"")
            }

            return @("${prefix}$Value")
        }

        if ($Value -is [bool]) {
            return @("${prefix}$($Value.ToString().ToLowerInvariant())")
        }

        if ($Value -is [int] -or $Value -is [long] -or $Value -is [double] -or $Value -is [decimal]) {
            return @("${prefix}$Value")
        }

        if ($Value -is [System.Collections.IDictionary]) {
            $lines = New-Object System.Collections.Generic.List[string]
            foreach ($entry in $Value.GetEnumerator() | Sort-Object Key) {
                if ($null -eq $entry.Value) {
                    $lines.Add("${prefix}$($entry.Key): null")
                    continue
                }

                if ($entry.Value -is [System.Collections.IDictionary] -or ($entry.Value -is [System.Collections.IEnumerable] -and -not ($entry.Value -is [string]))) {
                    $lines.Add("${prefix}$($entry.Key):")
                    foreach ($line in Convert-ValueToYaml -Value $entry.Value -Indent ($Indent + $indentWidth)) {
                        $lines.Add($line)
                    }
                    continue
                }

                $scalar = @(Convert-ValueToYaml -Value $entry.Value -Indent 0)
                $lines.Add("${prefix}$($entry.Key): $($scalar[0].ToString().TrimStart())")
            }
            return $lines
        }

        if ($Value -is [System.Collections.IEnumerable] -and -not ($Value -is [string])) {
            $lines = New-Object System.Collections.Generic.List[string]
            $items = @($Value)
            if ($items.Count -eq 0) {
                $lines.Add("${prefix}[]")
                return $lines
            }

            foreach ($item in $items) {
                if ($null -eq $item) {
                    $lines.Add("${prefix}- null")
                    continue
                }

                if ($item -is [System.Collections.IDictionary] -or ($item -is [System.Collections.IEnumerable] -and -not ($item -is [string]))) {
                    $lines.Add("${prefix}-")
                    foreach ($line in Convert-ValueToYaml -Value $item -Indent ($Indent + $indentWidth)) {
                        $lines.Add($line)
                    }
                    continue
                }

                $scalar = @(Convert-ValueToYaml -Value $item -Indent 0)
                $lines.Add("${prefix}- $($scalar[0].ToString().TrimStart())")
            }
            return $lines
        }

        return @("${prefix}$Value")
    }

    $ordered = $json | ConvertFrom-Json -AsHashtable
    (Convert-ValueToYaml -Value $ordered) -join [Environment]::NewLine
}
