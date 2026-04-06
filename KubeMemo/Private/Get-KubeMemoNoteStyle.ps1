function Get-KubeMemoNoteStyle {
    [CmdletBinding()]
    param(
        [string]$NoteType,
        [string]$Severity,
        [switch]$Runtime
    )

    $typeColor = switch ($NoteType) {
        'warning' { '38;5;214' }
        'temporary-warning' { '38;5;214' }
        'incident' { '38;5;203' }
        'activity' { '38;5;81' }
        'runbook' { '38;5;75' }
        'ownership' { '38;5;71' }
        'maintenance' { '38;5;179' }
        'handover' { '38;5;176' }
        'suppression' { '38;5;244' }
        'temporary-suppression' { '38;5;244' }
        default {
            if ($Runtime) { '38;5;117' } else { '38;5;153' }
        }
    }

    $paperColor = if ($Runtime) { '48;5;236;38;5;255' } else { '48;5;230;38;5;16' }
    $headerColor = if ($Runtime) { '48;5;214;38;5;16' } else { '48;5;110;38;5;16' }

    $severityColor = switch ($Severity) {
        'critical' { '48;5;124;38;5;231' }
        'error' { '48;5;160;38;5;231' }
        'warning' { '48;5;214;38;5;16' }
        'info' { '48;5;31;38;5;231' }
        default { '48;5;239;38;5;255' }
    }

    [pscustomobject]@{
        TypeColor = $typeColor
        PaperColor = $paperColor
        HeaderColor = $headerColor
        SeverityColor = $severityColor
    }
}
