function Merge-KubeMemoActivityNote {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [pscustomobject]$ExistingNote,

        [Parameter(Mandatory)]
        [hashtable]$Activity,

        [timespan]$Window = ([timespan]::FromSeconds(60))
    )

    if (-not $ExistingNote.Activity) {
        return $false
    }

    $sameShape = $ExistingNote.Activity.action -eq $Activity.action `
        -and $ExistingNote.Activity.fieldPath -eq $Activity.fieldPath `
        -and $ExistingNote.Activity.oldValue -eq $Activity.oldValue `
        -and $ExistingNote.Activity.newValue -eq $Activity.newValue

    if (-not $sameShape) {
        return $false
    }

    $existingCreatedAt = if ($ExistingNote.CreatedAt) { [datetime]$ExistingNote.CreatedAt } else { [datetime]::UtcNow.AddYears(-10) }
    if (([datetime]::UtcNow - $existingCreatedAt) -gt $Window) {
        return $false
    }

    return $true
}
