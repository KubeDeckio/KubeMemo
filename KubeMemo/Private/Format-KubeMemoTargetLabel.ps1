function Format-KubeMemoTargetLabel {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [psobject]$Note
    )

    if ($Note.TargetMode -eq 'app') {
        return "app/$($Note.AppName)" + $(if ($Note.AppInstance) { "/$($Note.AppInstance)" } else { '' })
    }

    if ($Note.TargetMode -eq 'namespace') {
        return "namespace/$($Note.Namespace)"
    }

    return "$($Note.Kind)/$($Note.Namespace)/$($Note.Name)"
}
