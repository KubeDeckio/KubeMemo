function Set-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory)]
        [string]$Id,
        [string]$Title,
        [string]$Summary,
        [string]$Content,
        [string[]]$Tags,
        [datetime]$ExpiresAt,
        [switch]$Runtime,
        [switch]$AnnotateResource,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [string]$OutputPath
    )

    $notes = Get-KubeMemo -IncludeRuntime:$Runtime
    $note = $notes | Where-Object Id -eq $Id | Select-Object -First 1
    if (-not $note) {
        throw "KubeMemo note '$Id' was not found."
    }

    if ($note.StoreType -eq 'Durable' -and -not (Test-KubeMemoDurableWriteAllowed)) {
        if (-not $OutputPath) {
            throw 'Durable note edits in GitOps mode require -OutputPath.'
        }
    }

    if ($note.StoreType -eq 'Runtime' -and -not (Test-KubeMemoRuntimeWriteAllowed -Namespace $RuntimeNamespace)) {
        throw 'Runtime note edits are not allowed because the runtime store is unavailable or unsafe.'
    }

    if ($Title) { $note.Title = $Title }
    if ($PSBoundParameters.ContainsKey('Summary')) { $note.Summary = $Summary }
    if ($PSBoundParameters.ContainsKey('Content')) { $note.Content = $Content }
    if ($PSBoundParameters.ContainsKey('Tags')) { $note.Tags = $Tags }
    if ($PSBoundParameters.ContainsKey('ExpiresAt')) { $note.ExpiresAt = $ExpiresAt.ToUniversalTime().ToString('o') }
    $note.UpdatedBy = Get-KubeMemoActor

    $spec = ConvertFrom-KubeMemoObject -InputObject $note
    $resource = if ($note.StoreType -eq 'Runtime') {
        ConvertTo-KubeMemoRuntimeNoteResource -ResourceName $note.Id -Namespace $RuntimeNamespace -Spec $spec
    }
    else {
        ConvertTo-KubeMemoNoteResource -ResourceName $note.Id -Namespace $note.RawResource.metadata.namespace -Spec $spec
    }

    $json = $resource | ConvertTo-Json -Depth 10
    if ($OutputPath) {
        if ($PSCmdlet.ShouldProcess($OutputPath, 'Write KubeMemo manifest')) {
            Set-Content -Path $OutputPath -Value $json
        }
        return Get-Item -Path $OutputPath
    }

    if ($PSCmdlet.ShouldProcess($note.Id, 'Update KubeMemo note')) {
        Invoke-KubeMemoKubectl -Arguments @('apply', '-f', '-') -InputObject $json | Out-Null
        if ($AnnotateResource -and $note.TargetMode -eq 'resource') {
            Set-KubeMemoResourceAnnotations -Kind $note.Kind -Name $note.Name -Namespace $note.Namespace -Summary $note.Summary -NoteRef $note.Id -RuntimeEnabled:($note.StoreType -eq 'Runtime') -WhatIf:$WhatIfPreference -Confirm:$false
        }
    }

    ConvertTo-KubeMemoObject -Resource ($json | ConvertFrom-Json) -StoreType $note.StoreType
}
