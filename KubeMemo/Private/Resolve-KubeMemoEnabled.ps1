function Resolve-KubeMemoEnabled {
    [CmdletBinding()]
    param(
        [string]$Kind,
        [string]$Namespace,
        [string]$Name,
        [string]$AppName,
        [string]$AppInstance,
        [switch]$IncludeRuntime,
        [hashtable]$Annotations
    )

    if ($Annotations) {
        if ($Annotations['notes.kubememo.io/enabled'] -eq 'true') {
            return $true
        }
        if ($Annotations['notes.kubememo.io/runtime-enabled'] -eq 'true' -and $IncludeRuntime) {
            return $true
        }
    }

    $notes = Get-KubeMemo -IncludeRuntime:$IncludeRuntime
    foreach ($note in $notes) {
        if ($Kind -and $note.Kind -eq $Kind -and $Namespace -and $note.Namespace -eq $Namespace -and $Name -and $note.Name -eq $Name) {
            return $true
        }

        if ($Namespace -and $note.TargetMode -eq 'namespace' -and $note.Namespace -eq $Namespace) {
            return $true
        }

        if ($AppName -and $note.AppName -eq $AppName) {
            if (-not $AppInstance -or $note.AppInstance -eq $AppInstance) {
                return $true
            }
        }
    }

    return $false
}
