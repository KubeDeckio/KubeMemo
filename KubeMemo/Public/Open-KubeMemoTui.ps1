function Open-KubeMemoTui {
    [CmdletBinding()]
    param(
        [switch]$IncludeRuntime,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [string[]]$Namespace,
        [int]$AutoRefreshSeconds = 3
    )

    $selectedIndex = 0
    $detailScroll = 0
    $namespaceFilter = @($Namespace | Where-Object { $_ } | Select-Object -Unique)
    $kindFilter = ''
    $textFilter = ''
    $viewMode = 'all'
    $statusMessage = 'Ready'

    function Get-VisibleWidth {
        param(
            [AllowNull()]
            [string]$Text
        )

        if ($null -eq $Text) {
            return 0
        }

        $ansiPattern = [regex]"`e\[[0-9;]*m"
        return $ansiPattern.Replace($Text, '').Length
    }

    function Fit-VisibleText {
        param(
            [AllowNull()]
            [string]$Text,
            [int]$Width
        )

        if ($Width -le 0) {
            return ''
        }

        if ([string]::IsNullOrEmpty($Text)) {
            return ''
        }

        if ($Text.Length -le $Width) {
            return $Text
        }

        if ($Width -eq 1) {
            return '…'
        }

        return $Text.Substring(0, $Width - 1) + '…'
    }

    function Pad-VisibleText {
        param(
            [AllowNull()]
            [string]$Text,
            [int]$Width
        )

        $value = if ($null -eq $Text) { '' } else { $Text }
        $visibleWidth = Get-VisibleWidth -Text $value
        if ($visibleWidth -lt $Width) {
            return $value + (' ' * ($Width - $visibleWidth))
        }

        return $value
    }

    while ($true) {
        $windowState = Ensure-KubeMemoWindowSize
        $access = Get-KubeMemoAccessScope -RuntimeNamespace $RuntimeNamespace
        $state = Get-KubeMemoTuiState -Namespaces $namespaceFilter -IncludeRuntime:$IncludeRuntime -RuntimeNamespace $RuntimeNamespace -Kind $kindFilter -Text $textFilter
        $notes = @($state.Notes)
        if ($viewMode -eq 'memo') {
            $notes = @($notes | Where-Object StoreType -eq 'Durable')
        }
        elseif ($viewMode -eq 'runtimememo') {
            $notes = @($notes | Where-Object StoreType -eq 'Runtime')
        }

        $visibleNamespaces = @($state.VisibleNamespaces)
        $visibleKinds = @($state.VisibleKinds)

        $uiWidth = 96
        $uiHeight = 32
        try {
            if ($Host.UI.RawUI.WindowSize.Width -gt 0) {
                $uiWidth = $Host.UI.RawUI.WindowSize.Width
            }
            if ($Host.UI.RawUI.WindowSize.Height -gt 0) {
                $uiHeight = $Host.UI.RawUI.WindowSize.Height
            }
        }
        catch {
        }

        $boardWidth = [Math]::Max(90, $uiWidth - 2)
        $splitPane = $boardWidth -ge 118
        $listPaneWidth = if ($splitPane) { [Math]::Floor($boardWidth * 0.50) } else { $boardWidth }
        $detailPaneWidth = if ($splitPane) { $boardWidth - $listPaneWidth - 3 } else { $boardWidth }
        $namespaceWidth = [Math]::Max(8, [Math]::Min(12, [int](($listPaneWidth - 16) * 0.22)))
        $titleWidth = [Math]::Max(28, $listPaneWidth - (23 + $namespaceWidth))
        $maxDetailLines = if ($splitPane) { [Math]::Max(14, $uiHeight - 16) } else { [Math]::Max(10, $uiHeight - 24) }

        if ($notes.Count -gt 0 -and $selectedIndex -ge $notes.Count) {
            $selectedIndex = $notes.Count - 1
        }
        if ($selectedIndex -lt 0) {
            $selectedIndex = 0
        }
        if ($notes.Count -eq 0) {
            $detailScroll = 0
        }

        Clear-Host
        $scopeText = if ($namespaceFilter.Count -gt 0) { $namespaceFilter -join ', ' } else { 'all accessible namespaces' }
        $kindText = if ($kindFilter) { $kindFilter } else { 'all kinds' }
        $textText = if ($textFilter) { $textFilter } else { 'none' }
        $viewText = switch ($viewMode) {
            'memo' { 'memo' }
            'runtimememo' { 'runtimememo' }
            default { 'all' }
        }
        $accessText = if ($access.DurableScope -like 'namespace:*') {
            "read scope limited to $($access.DurableScope.Split(':')[1]) by RBAC"
        }
        elseif ($access.DurableScope -eq 'cluster') {
            'cluster-wide durable read'
        }
        else {
            'durable memo read unavailable'
        }
        if ($IncludeRuntime -and -not $access.RuntimeReadable) {
            $accessText += " | runtime namespace $RuntimeNamespace not readable"
        }
        if (-not $windowState.Supported) {
            $accessText += ' | host cannot auto-resize'
        }

        Show-KubeMemoBanner
        Write-Host (Format-KubeMemoAnsi -Text (" SCOPE  $scopeText ") -Code '30;48;5;153')
        Write-Host (Format-KubeMemoAnsi -Text (" ACCESS  $accessText ") -Code '30;48;5;186')
        Write-Host (Format-KubeMemoAnsi -Text (" FILTERS  view=$viewText  ns=$scopeText  kind=$kindText  text=$textText ") -Code '30;48;5;110')
        Write-Host ''

        $listLines = New-Object System.Collections.Generic.List[string]
        $listLines.Add((Format-KubeMemoAnsi -Text (Pad-VisibleText -Text (" LIST  {0} memo(s)" -f $notes.Count) -Width $listPaneWidth) -Code '30;48;5;45'))
        $listLines.Add((Format-KubeMemoAnsi -Text ('─' * $listPaneWidth) -Code '38;5;45'))
        $headerText = " #  SRC   TYPE        {0,-$namespaceWidth} {1}" -f 'NS', 'TITLE'
        $listLines.Add((Format-KubeMemoAnsi -Text (Pad-VisibleText -Text $headerText -Width $listPaneWidth) -Code '38;5;244'))
        $listLines.Add((Format-KubeMemoAnsi -Text ('─' * $listPaneWidth) -Code '38;5;45'))

        if ($notes.Count -eq 0) {
            $emptyLine = Pad-VisibleText -Text ' No memos found for the current scope.' -Width $listPaneWidth
            $listLines.Add((Format-KubeMemoAnsi -Text $emptyLine -Code '38;5;214'))
        }
        else {
            for ($index = 0; $index -lt $notes.Count; $index++) {
                $note = $notes[$index]
                $store = if ($note.StoreType -eq 'Runtime') { 'RUN' } else { 'MEM' }
                $title = Fit-VisibleText -Text $note.Title -Width $titleWidth
                $namespaceText = if ($note.Namespace) { $note.Namespace } else { '-' }
                $typeText = $note.NoteType.ToUpperInvariant()
                if ($typeText.Length -gt 10) {
                    $typeText = $typeText.Substring(0, 9) + '…'
                }
                $row = " {0} {1,-2} {2,-4} {3,-10} {4,-$namespaceWidth} {5}" -f $(if ($index -eq $selectedIndex) { '>' } else { ' ' }), ($index + 1), $store, $typeText, $namespaceText, $title
                $row = Pad-VisibleText -Text $row -Width $listPaneWidth
                if ($index -eq $selectedIndex) {
                    $listLines.Add((Format-KubeMemoAnsi -Text $row -Code '1;30;48;5;153'))
                }
                else {
                    $code = if ($note.StoreType -eq 'Runtime') { '38;5;220' } else { '38;5;120' }
                    $listLines.Add((Format-KubeMemoAnsi -Text $row -Code $code))
                }
            }
        }

        $detailLines = New-Object System.Collections.Generic.List[string]
        $detailTitle = if ($notes.Count -gt 0) { " DETAIL  $(Format-KubeMemoTargetLabel -Note $notes[$selectedIndex]) " } else { ' DETAIL ' }
        $detailTitle = Fit-VisibleText -Text $detailTitle -Width $detailPaneWidth
        $detailLines.Add((Format-KubeMemoAnsi -Text (Pad-VisibleText -Text $detailTitle -Width $detailPaneWidth) -Code '30;48;5;45'))
        $detailLines.Add((Format-KubeMemoAnsi -Text ('─' * $detailPaneWidth) -Code '38;5;45'))
        if ($notes.Count -gt 0) {
            $detailText = Write-KubeMemoCard -Notes @($notes[$selectedIndex]) -PassThru -NoHeader -Width ([Math]::Max(24, $detailPaneWidth - 4))
            $allDetailLines = @($detailText -split "`r?`n")
            $maxDetailScroll = [Math]::Max(0, $allDetailLines.Count - $maxDetailLines)
            if ($detailScroll -gt $maxDetailScroll) {
                $detailScroll = $maxDetailScroll
            }
            if ($detailScroll -lt 0) {
                $detailScroll = 0
            }
            foreach ($line in @($allDetailLines | Select-Object -Skip $detailScroll -First $maxDetailLines)) {
                $detailLines.Add((Pad-VisibleText -Text $line -Width $detailPaneWidth))
            }
            if ($maxDetailScroll -gt 0) {
                $scrollText = Pad-VisibleText -Text (" detail scroll {0}/{1}" -f $detailScroll, $maxDetailScroll) -Width $detailPaneWidth
                $detailLines.Add((Format-KubeMemoAnsi -Text $scrollText -Code '38;5;244'))
            }
        }
        else {
            $detailLines.Add((Format-KubeMemoAnsi -Text (Pad-VisibleText -Text ' Nothing selected.' -Width $detailPaneWidth) -Code '38;5;244'))
        }

        if ($splitPane) {
            $rows = [Math]::Max($listLines.Count, $detailLines.Count)
            for ($i = 0; $i -lt $rows; $i++) {
                $left = if ($i -lt $listLines.Count) { $listLines[$i] } else { '' }
                $right = if ($i -lt $detailLines.Count) { $detailLines[$i] } else { '' }
                $leftPad = Pad-VisibleText -Text $left -Width $listPaneWidth
                Write-Host ($leftPad + ' │ ' + $right)
            }
        }
        else {
            foreach ($line in $listLines) {
                Write-Host $line
            }
            Write-Host ''
            foreach ($line in $detailLines) {
                Write-Host $line
            }
        }

        Write-Host ''
        $statusBar = " STATUS  $statusMessage "
        Write-Host (Format-KubeMemoAnsi -Text $statusBar -Code '30;48;5;109')
        $visibleKindsText = if ($visibleKinds.Count -gt 0) { $visibleKinds -join ',' } else { 'none' }
        Write-Host (Format-KubeMemoAnsi -Text (" visible kinds: $visibleKindsText ") -Code '38;5;244')
        Write-Host (Format-KubeMemoAnsi -Text (" [Arrows]/[j][k] move  [PgUp]/[PgDn] or [u][d] scroll  [/] text  [:] view  [f] ns  [c] kind  [a] add  [r] refresh  [q] quit  auto ${AutoRefreshSeconds}s ") -Code '38;5;45')

        $key = $null
        $pollIntervalMs = 200
        $waitedMs = 0
        while (-not $key) {
            try {
                if ([Console]::KeyAvailable) {
                    $key = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
                    break
                }
            }
            catch {
                $key = $Host.UI.RawUI.ReadKey('NoEcho,IncludeKeyDown')
                break
            }

            Start-Sleep -Milliseconds $pollIntervalMs
            $waitedMs += $pollIntervalMs
            if ($AutoRefreshSeconds -gt 0 -and $waitedMs -ge ($AutoRefreshSeconds * 1000)) {
                $statusMessage = "Auto-refreshed $(Get-Date -Format 'HH:mm:ss')"
                continue 2
            }
        }
        $statusMessage = 'Ready'

        switch ($key.VirtualKeyCode) {
            38 {
                if ($selectedIndex -gt 0) {
                    $selectedIndex--
                    $detailScroll = 0
                }
                continue
            }
            40 {
                if ($notes.Count -gt 0) {
                    $selectedIndex = [Math]::Min($selectedIndex + 1, $notes.Count - 1)
                    $detailScroll = 0
                }
                continue
            }
            33 {
                if ($detailScroll -gt 0) {
                    $detailScroll = [Math]::Max(0, $detailScroll - 6)
                    $statusMessage = "Detail scrolled to line $detailScroll."
                }
                continue
            }
            34 {
                if ($notes.Count -gt 0) {
                    $detailText = Write-KubeMemoCard -Notes @($notes[$selectedIndex]) -PassThru -NoHeader -Width ([Math]::Max(24, $detailPaneWidth - 4))
                    $detailLineCount = @($detailText -split "`r?`n").Count
                    $maxDetailScroll = [Math]::Max(0, $detailLineCount - $maxDetailLines)
                    if ($detailScroll -lt $maxDetailScroll) {
                        $detailScroll = [Math]::Min($maxDetailScroll, $detailScroll + 6)
                        $statusMessage = "Detail scrolled to line $detailScroll."
                    }
                }
                continue
            }
        }

        $action = $key.Character.ToString()
        switch ($action) {
            'q' { return }
            'r' { continue }
            'j' {
                if ($notes.Count -gt 0) {
                    $selectedIndex = [Math]::Min($selectedIndex + 1, $notes.Count - 1)
                    $detailScroll = 0
                }
            }
            'k' {
                if ($selectedIndex -gt 0) {
                    $selectedIndex--
                    $detailScroll = 0
                }
            }
            'c' {
                Write-Host ''
                $input = Read-Host "Kind filter (blank to clear, example: Deployment)"
                if ([string]::IsNullOrWhiteSpace($input)) {
                    $kindFilter = ''
                    $selectedIndex = 0
                    $detailScroll = 0
                    $statusMessage = 'Kind filter cleared.'
                }
                else {
                    $kindFilter = $input.Trim()
                    $selectedIndex = 0
                    $detailScroll = 0
                    $statusMessage = "Kind filter set to $kindFilter."
                }
            }
            'g' {
                $selectedIndex = 0
                $detailScroll = 0
            }
            'G' {
                if ($notes.Count -gt 0) {
                    $selectedIndex = $notes.Count - 1
                    $detailScroll = 0
                }
            }
            'u' {
                if ($detailScroll -gt 0) {
                    $detailScroll = [Math]::Max(0, $detailScroll - 3)
                    $statusMessage = "Detail scrolled to line $detailScroll."
                }
            }
            'd' {
                if ($notes.Count -gt 0) {
                    $detailText = Write-KubeMemoCard -Notes @($notes[$selectedIndex]) -PassThru -NoHeader -Width ([Math]::Max(24, $detailPaneWidth - 4))
                    $detailLineCount = @($detailText -split "`r?`n").Count
                    $maxDetailScroll = [Math]::Max(0, $detailLineCount - $maxDetailLines)
                    if ($detailScroll -lt $maxDetailScroll) {
                        $detailScroll = [Math]::Min($maxDetailScroll, $detailScroll + 3)
                        $statusMessage = "Detail scrolled to line $detailScroll."
                    }
                }
            }
            '/' {
                Write-Host ''
                $input = Read-Host 'Filter text (blank to clear)'
                if ([string]::IsNullOrWhiteSpace($input)) {
                    $textFilter = ''
                    $selectedIndex = 0
                    $detailScroll = 0
                    $statusMessage = 'Text filter cleared.'
                }
                else {
                    $textFilter = $input.Trim()
                    $selectedIndex = 0
                    $detailScroll = 0
                    $statusMessage = 'Text filter updated.'
                }
            }
            ':' {
                Write-Host ''
                $command = (Read-Host 'Command').Trim().ToLowerInvariant()
                switch ($command) {
                    'memo' {
                        $viewMode = 'memo'
                        $selectedIndex = 0
                        $detailScroll = 0
                        $statusMessage = 'View changed to durable memos.'
                    }
                    'runtimememo' {
                        $viewMode = 'runtimememo'
                        $selectedIndex = 0
                        $detailScroll = 0
                        $statusMessage = 'View changed to runtime memos.'
                    }
                    'all' {
                        $viewMode = 'all'
                        $selectedIndex = 0
                        $detailScroll = 0
                        $statusMessage = 'View changed to all memos.'
                    }
                    '' {
                        $statusMessage = 'Command cancelled.'
                    }
                    default {
                        $statusMessage = "Unknown command: :$command"
                    }
                }
            }
            'f' {
                $input = Read-Host "Namespace filter (comma-separated, blank for current scope, '*' for all accessible)"
                if ([string]::IsNullOrWhiteSpace($input)) {
                    $statusMessage = "Scope unchanged: $scopeText"
                    continue
                }
                if ($input.Trim() -eq '*') {
                    $namespaceFilter = @()
                    $selectedIndex = 0
                    $detailScroll = 0
                    $statusMessage = 'Namespace filter cleared.'
                    continue
                }
                $namespaceFilter = @($input.Split(',') | ForEach-Object { $_.Trim() } | Where-Object { $_ } | Select-Object -Unique)
                $selectedIndex = 0
                $detailScroll = 0
                $statusMessage = 'Namespace filter updated.'
            }
            't' {
                $input = Read-Host 'Text filter (blank to clear, matches title/summary/content)'
                if ([string]::IsNullOrWhiteSpace($input)) {
                    $textFilter = ''
                    $selectedIndex = 0
                    $detailScroll = 0
                    $statusMessage = 'Text filter cleared.'
                }
                else {
                    $textFilter = $input.Trim()
                    $selectedIndex = 0
                    $detailScroll = 0
                    $statusMessage = 'Text filter updated.'
                }
            }
            'a' {
                $kind = Read-Host 'Kind'
                $namespace = Read-Host 'Namespace'
                $name = Read-Host 'Name'
                $title = Read-Host 'Title'
                $summary = Read-Host 'Summary'
                $content = Read-Host 'Content'
                $hours = Read-Host 'Expiry hours (default 24)'
                $expiryHours = 24
                [void][int]::TryParse($hours, [ref]$expiryHours)

                try {
                    New-KubeMemo -Temporary -Kind $kind -Namespace $namespace -Name $name -Title $title -Summary $summary -Content $content -NoteType incident -Severity info -ExpiresAt ([datetime]::UtcNow.AddHours($expiryHours)) | Out-Null
                    $statusMessage = 'Runtime memo created.'
                }
                catch {
                    $statusMessage = $_.Exception.Message
                }
            }
            '*' {
                $namespaceFilter = @()
                $selectedIndex = 0
                $detailScroll = 0
                $statusMessage = 'Showing all accessible namespaces.'
            }
            default {
                $numeric = 0
                if ([int]::TryParse($action, [ref]$numeric)) {
                    if ($numeric -ge 1 -and $numeric -le $notes.Count) {
                        $selectedIndex = $numeric - 1
                        $detailScroll = 0
                    }
                }
            }
        }
    }
}
