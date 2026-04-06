function Write-KubeMemoCard {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [psobject[]]$Notes,

        [switch]$PassThru,

        [switch]$NoColor,

        [switch]$NoHeader,

        [int]$Width = 78
    )

    $config = Get-KubeMemoConfigInternal
    $useAnsi = $config.Rendering.Ansi -and -not $NoColor
    $blocks = New-Object System.Collections.Generic.List[string]

    function Fit-PlainText {
        param(
            [AllowNull()]
            [string]$Text,
            [int]$MaxWidth
        )

        $value = if ($null -eq $Text) { '' } else { $Text }
        if ($MaxWidth -le 0) {
            return ''
        }

        if ($value.Length -le $MaxWidth) {
            return $value
        }

        if ($MaxWidth -eq 1) {
            return '…'
        }

        return $value.Substring(0, $MaxWidth - 1) + '…'
    }

    function Join-WrappedBlock {
        param(
            [string]$Label,
            [string]$Text
        )

        if ([string]::IsNullOrWhiteSpace($Text)) {
            return @()
        }

        $lines = New-Object System.Collections.Generic.List[string]
        $lines.Add($Label)
        foreach ($line in (Split-KubeMemoText -Text $Text -Width ($Width - 2))) {
            $lines.Add("  $line")
        }
        return $lines
    }

    function New-KubeMemoNoteCard {
        param(
            [string]$Title,
            [string[]]$Lines,
            [string]$BorderCode,
            [string]$TitleCode,
            [switch]$DisableColor
        )

        $contentWidth = [Math]::Max(24, $Width)
        $top = Format-KubeMemoAnsi -Text ('╭' + ('─' * ($contentWidth + 2)) + '╮') -Code $BorderCode -Disable:$DisableColor
        $bottom = Format-KubeMemoAnsi -Text ('╰' + ('─' * ($contentWidth + 2)) + '╯') -Code $BorderCode -Disable:$DisableColor
        $cardLines = New-Object System.Collections.Generic.List[string]
        $cardLines.Add($top)
        $titleText = Fit-PlainText -Text $Title -MaxWidth $contentWidth
        $titleInner = Format-KubeMemoAnsi -Text (" {0,-$contentWidth} " -f $titleText) -Code $TitleCode -Disable:$DisableColor
        $titleRow = (
            (Format-KubeMemoAnsi -Text '│' -Code $BorderCode -Disable:$DisableColor) +
            $titleInner +
            (Format-KubeMemoAnsi -Text '│' -Code $BorderCode -Disable:$DisableColor)
        )
        $cardLines.Add($titleRow)

        foreach ($line in $Lines) {
            $trimmed = Fit-PlainText -Text $line -MaxWidth $contentWidth
            $cardLines.Add((Format-KubeMemoAnsi -Text ("│ {0,-$contentWidth} │" -f $trimmed) -Code $BorderCode -Disable:$DisableColor))
        }

        $cardLines.Add($bottom)
        return ($cardLines -join [Environment]::NewLine)
    }

    $targetLabel = Format-KubeMemoTargetLabel -Note $Notes[0]
    if (-not $NoHeader) {
        $header = @(
            Format-KubeMemoAnsi -Text "KubeMemo  $targetLabel" -Code '1;38;5;45' -Disable:(-not $useAnsi)
            Format-KubeMemoAnsi -Text ('═' * [Math]::Max(72, ("KubeMemo  $targetLabel").Length)) -Code '38;5;60' -Disable:(-not $useAnsi)
        ) -join [Environment]::NewLine
        $blocks.Add($header)
    }

    foreach ($group in $Notes | Group-Object -Property StoreType) {
        $sectionTitle = if ($group.Name -eq 'Runtime') { 'Runtime Memos' } else { 'Durable Memos' }
        $sectionCode = if ($group.Name -eq 'Runtime') { '1;38;5;81' } else { '1;38;5;149' }
        $blocks.Add((Format-KubeMemoAnsi -Text $sectionTitle -Code $sectionCode -Disable:(-not $useAnsi)))

        foreach ($note in $group.Group) {
            $style = Get-KubeMemoNoteStyle -NoteType $note.NoteType -Severity $note.Severity -Runtime:($note.StoreType -eq 'Runtime')
            $subject = if ($note.Title) { $note.Title } else { $note.NoteType }
            $sourceLabel = if ($note.OwnerTeam -or $note.OwnerContact) {
                "$($note.OwnerTeam) $($note.OwnerContact)".Trim()
            }
            elseif ($note.StoreType -eq 'Durable' -and $note.CreatedBy) {
                $note.CreatedBy
            }
            elseif ($note.GitRepo -or $note.GitPath) {
                @($note.GitRepo, $note.GitPath | Where-Object { $_ }) -join ' '
            }
            elseif ($note.SourceType -eq 'git') {
                'git-managed note'
            }
            elseif ($note.StoreType -eq 'Durable' -and $note.SourceType -eq 'manual') {
                'curated memo'
            }
            elseif ($note.SourceGenerator) {
                "$($note.SourceType) via $($note.SourceGenerator)"
            }
            else {
                if ($note.StoreType -eq 'Runtime') { 'runtime store' } else { 'durable store' }
            }

            $metaLine = @("Target $(Format-KubeMemoTargetLabel -Note $note)")
            if ($note.StoreType -eq 'Runtime') {
                $metaLine += "Source $sourceLabel"
            }
            else {
                $metaLine += "Owner $sourceLabel"
            }
            if ($note.Tags -and $note.Tags.Count -gt 0) {
                $metaLine += ("Tags " + ($note.Tags -join ', '))
            }

            $bodyLines = New-Object System.Collections.Generic.List[string]
            $metaHeader = @($note.NoteType)
            if ($note.Severity) {
                $metaHeader += $note.Severity
            }
            if ($note.StoreType -eq 'Runtime' -and $note.ExpiresAt) {
                $metaHeader += ("expires " + (Format-KubeMemoRelativeTime -Timestamp $note.ExpiresAt))
            }
            $bodyLines.Add((($metaHeader | ForEach-Object { $_.ToString().Trim() } | Where-Object { $_ }) -join '  |  ').ToUpperInvariant())
            $bodyLines.Add(($metaLine -join '   |   '))

            if ($note.StoreType -eq 'Runtime') {
                $bodyLines.Add('Temporary live operational context.')
            }

            foreach ($line in (Join-WrappedBlock -Label 'Summary' -Text $note.Summary)) {
                $bodyLines.Add($line)
            }

            $detailLabel = switch ($note.NoteType) {
                'runbook' { 'Runbook' }
                'warning' { 'Guidance' }
                'incident' { 'Notes' }
                'activity' { 'Notes' }
                default { 'Details' }
            }
            $detailText = Render-KubeMemoMarkdown -Content $note.Content
            foreach ($line in (Join-WrappedBlock -Label $detailLabel -Text $detailText)) {
                $bodyLines.Add($line)
            }

            if ($note.Activity) {
                $activityText = "Activity $($note.Activity.action)  $($note.Activity.oldValue) -> $($note.Activity.newValue)"
                $bodyLines.Add($activityText)
            }

            $borderCode = if ($note.StoreType -eq 'Runtime') { '38;5;214' } else { '38;5;110' }
            $blocks.Add((New-KubeMemoNoteCard -Title $subject -Lines $bodyLines -BorderCode $borderCode -TitleCode $style.HeaderColor -DisableColor:(-not $useAnsi)))
        }
    }

    $rendered = ($blocks -join ([Environment]::NewLine + [Environment]::NewLine)).Trim()
    if ($PassThru) {
        return $rendered
    }

    Write-Host $rendered
}
