function ConvertTo-KubeMemoObject {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [psobject]$Resource,

        [Parameter(Mandatory)]
        [ValidateSet('Durable', 'Runtime')]
        [string]$StoreType
    )

    $spec = $Resource.spec
    $metadata = $Resource.metadata
    $source = $spec.source
    $activity = $spec.activity
    $owner = $spec.owner
    $target = $spec.target

    [pscustomobject]@{
        Id = $metadata.name
        StoreType = $StoreType
        Title = $spec.title
        Summary = $spec.summary
        Content = $spec.content
        Format = $spec.format
        NoteType = $spec.noteType
        Temporary = [bool]$spec.temporary
        Severity = $spec.severity
        OwnerTeam = $owner.team
        OwnerContact = $owner.contact
        Tags = @($spec.tags)
        Links = @($spec.links)
        TargetMode = $target.mode
        ApiVersion = $target.apiVersion
        Kind = $target.kind
        Namespace = $target.namespace
        Name = $target.name
        AppName = $target.appRef.name
        AppInstance = $target.appRef.instance
        ValidFrom = $spec.validFrom
        ExpiresAt = $spec.expiresAt
        CreatedAt = if ($spec.createdAt) { $spec.createdAt } else { $metadata.creationTimestamp }
        UpdatedAt = $metadata.creationTimestamp
        CreatedBy = $spec.createdBy
        UpdatedBy = $spec.updatedBy
        SourceType = $source.type
        SourceGenerator = $source.generator
        Confidence = $source.confidence
        GitRepo = $source.git.repo
        GitPath = $source.git.path
        GitRevision = $source.git.revision
        Activity = $activity
        RawResource = $Resource
    }
}
