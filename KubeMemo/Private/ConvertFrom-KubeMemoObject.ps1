function ConvertFrom-KubeMemoObject {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [pscustomobject]$InputObject
    )

    return @{
        title = $InputObject.Title
        summary = $InputObject.Summary
        content = $InputObject.Content
        format = if ($InputObject.Format) { $InputObject.Format } else { 'markdown' }
        noteType = $InputObject.NoteType
        temporary = $InputObject.Temporary
        severity = $InputObject.Severity
        target = @{
            mode = $InputObject.TargetMode
            apiVersion = $InputObject.ApiVersion
            kind = $InputObject.Kind
            namespace = $InputObject.Namespace
            name = $InputObject.Name
            appRef = @{
                name = $InputObject.AppName
                instance = $InputObject.AppInstance
            }
        }
        owner = @{
            team = $InputObject.OwnerTeam
            contact = $InputObject.OwnerContact
        }
        tags = @($InputObject.Tags | Where-Object { $null -ne $_ -and $_ -ne '' })
        links = @($InputObject.Links | Where-Object { $null -ne $_ })
        validFrom = $InputObject.ValidFrom
        expiresAt = $InputObject.ExpiresAt
        createdBy = $InputObject.CreatedBy
        updatedBy = $InputObject.UpdatedBy
    }
}
