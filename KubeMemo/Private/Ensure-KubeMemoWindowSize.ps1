function Ensure-KubeMemoWindowSize {
    [CmdletBinding()]
    param(
        [int]$MinimumWidth = 124,
        [int]$MinimumHeight = 34
    )

    try {
        $rawUi = $Host.UI.RawUI
        $windowSize = $rawUi.WindowSize
        $bufferSize = $rawUi.BufferSize
        $changed = $false

        if ($bufferSize.Width -lt $MinimumWidth) {
            $bufferSize.Width = $MinimumWidth
            $rawUi.BufferSize = $bufferSize
            $changed = $true
        }
        if ($bufferSize.Height -lt $MinimumHeight) {
            $bufferSize.Height = $MinimumHeight
            $rawUi.BufferSize = $bufferSize
            $changed = $true
        }

        if ($windowSize.Width -lt $MinimumWidth -or $windowSize.Height -lt $MinimumHeight) {
            $rawUi.WindowSize = New-Object System.Management.Automation.Host.Size(
                [Math]::Max($windowSize.Width, $MinimumWidth),
                [Math]::Max($windowSize.Height, $MinimumHeight)
            )
            $changed = $true
        }

        return [pscustomobject]@{
            Supported = $true
            Changed = $changed
            Width = $rawUi.WindowSize.Width
            Height = $rawUi.WindowSize.Height
        }
    }
    catch {
        return [pscustomobject]@{
            Supported = $false
            Changed = $false
            Width = $null
            Height = $null
        }
    }
}
