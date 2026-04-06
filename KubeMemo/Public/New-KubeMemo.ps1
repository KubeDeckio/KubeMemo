function New-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, DefaultParameterSetName = 'Resource', ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory)]
        [string]$Title,

        [string]$Summary,
        [string]$Content,
        [string]$Format = 'markdown',
        [string]$NoteType = 'advisory',
        [string]$Severity = 'info',
        [string]$OwnerTeam,
        [string]$OwnerContact,
        [string[]]$Tags,
        [datetime]$ExpiresAt,
        [switch]$Temporary,
        [switch]$AnnotateResource,
        [string]$OutputPath,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,

        [Parameter(ParameterSetName = 'Resource', Mandatory)]
        [string]$Kind,
        [Parameter(ParameterSetName = 'Resource', Mandatory)]
        [string]$Name,
        [Parameter(ParameterSetName = 'Resource')]
        [string]$ApiVersion = 'v1',
        [Parameter(ParameterSetName = 'Resource')]
        [string]$Namespace,
        [Parameter(ParameterSetName = 'Namespace', Mandatory)]
        [string]$TargetNamespace,
        [Parameter(ParameterSetName = 'App', Mandatory)]
        [string]$AppName,
        [Parameter(ParameterSetName = 'App')]
        [string]$AppInstance
    )

    $targetArgs = @{}
    foreach ($parameterName in 'Kind', 'Name', 'ApiVersion', 'Namespace', 'TargetNamespace', 'AppName', 'AppInstance') {
        if ($PSBoundParameters.ContainsKey($parameterName)) {
            $targetArgs[$parameterName] = $PSBoundParameters[$parameterName]
        }
    }
    $target = Resolve-KubeMemoTarget @targetArgs
    $resourceName = (($Title.ToLowerInvariant() -replace '[^a-z0-9]+', '-') -replace '(^-|-$)', '')
    if (-not $resourceName) {
        $resourceName = "kubememo-{0}" -f ([guid]::NewGuid().ToString('N').Substring(0, 8))
    }
    $actor = Get-KubeMemoActor
    $timestamp = [datetime]::UtcNow.ToString('o')

    $spec = @{
        title = $Title
        summary = $Summary
        content = $Content
        format = $Format
        noteType = $NoteType
        severity = $Severity
        target = $target
        owner = @{
            team = $OwnerTeam
            contact = $OwnerContact
        }
        tags = @($Tags | Where-Object { $null -ne $_ -and $_ -ne '' })
        links = @()
        validFrom = $timestamp
        expiresAt = if ($PSBoundParameters.ContainsKey('ExpiresAt')) { $ExpiresAt.ToUniversalTime().ToString('o') } else { $null }
        createdBy = $actor
        updatedBy = $actor
    }

    if ($Temporary) {
        if (-not (Test-KubeMemoRuntimeWriteAllowed -Namespace $RuntimeNamespace)) {
            throw 'Runtime note creation is not allowed because the runtime store is unavailable or unsafe in GitOps mode.'
        }

        $spec.temporary = $true
        $spec.createdAt = $timestamp
        $spec.source = @{
            type = 'manual'
            generator = 'New-KubeMemo'
            confidence = 'high'
        }
        $resource = ConvertTo-KubeMemoRuntimeNoteResource -ResourceName $resourceName -Namespace $RuntimeNamespace -Spec $spec
    }
    else {
        $gitOps = Resolve-KubeMemoGitOpsState
        $spec.source = @{
            type = if ($gitOps.Enabled) { 'git' } else { 'manual' }
        }
        $resourceNamespace = if ($Namespace) { $Namespace } elseif ($TargetNamespace) { $TargetNamespace } else { 'default' }
        $resource = ConvertTo-KubeMemoNoteResource -ResourceName $resourceName -Namespace $resourceNamespace -Spec $spec
    }

    $json = $resource | ConvertTo-Json -Depth 10
    $yaml = ConvertTo-KubeMemoYaml -Resource $resource

    if (-not $Temporary -and -not (Test-KubeMemoDurableWriteAllowed)) {
        if ($OutputPath) {
            if ($PSCmdlet.ShouldProcess($OutputPath, 'Write durable note manifest for GitOps')) {
                Set-Content -Path $OutputPath -Value $yaml
            }
            return Get-Item -Path $OutputPath
        }

        return $yaml
    }

    if ($PSCmdlet.ShouldProcess($resource.metadata.name, 'Create KubeMemo note')) {
        Invoke-KubeMemoKubectl -Arguments @('apply', '-f', '-') -InputObject $json | Out-Null
        if ($AnnotateResource -and $target.mode -eq 'resource') {
            Set-KubeMemoResourceAnnotations -Kind $target.kind -Name $target.name -Namespace $target.namespace -Summary $Summary -NoteRef $resourceName -RuntimeEnabled:$Temporary -WhatIf:$WhatIfPreference -Confirm:$false
        }
    }

    ConvertTo-KubeMemoObject -Resource ($json | ConvertFrom-Json) -StoreType $(if ($Temporary) { 'Runtime' } else { 'Durable' })
}
