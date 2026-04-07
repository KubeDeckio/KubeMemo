function Install-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [switch]$DurableOnly,
        [switch]$EnableRuntimeStore,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [switch]$GitOpsAware,
        [switch]$InstallRbac
    )

    if ($PSCmdlet.ShouldProcess('cluster', 'Install KubeMemo prerequisites')) {
        $args = @('install', '--output', 'json', '--runtime-namespace', $RuntimeNamespace)
        if ($DurableOnly) { $args += '--durable-only' }
        if ($EnableRuntimeStore) { $args += '--enable-runtime-store' }
        if ($InstallRbac) { $args += '--install-rbac' }
        if ($GitOpsAware) { $args += '--gitops-aware' }
        Invoke-KubeMemoBinary -Arguments $args -ParseJson
    }
}

function Uninstall-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [switch]$RuntimeOnly,
        [switch]$RemoveData,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    if ($PSCmdlet.ShouldProcess('cluster', 'Uninstall KubeMemo prerequisites')) {
        $args = @('uninstall', '--output', 'json', '--runtime-namespace', $RuntimeNamespace)
        if ($RuntimeOnly) { $args += '--runtime-only' }
        if ($RemoveData) { $args += '--remove-data' }
        Invoke-KubeMemoBinary -Arguments $args -ParseJson
    }
}

function Update-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [switch]$IncludeRbac,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [switch]$GitOpsAware
    )

    if ($PSCmdlet.ShouldProcess('cluster', 'Update KubeMemo prerequisites')) {
        $args = @('update', '--output', 'json', '--runtime-namespace', $RuntimeNamespace)
        if ($IncludeRbac) { $args += '--include-rbac' }
        if ($GitOpsAware) { $args += '--gitops-aware' }
        Invoke-KubeMemoBinary -Arguments $args -ParseJson
    }
}

function Test-KubeMemoInstallation {
    [CmdletBinding()]
    param(
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    Invoke-KubeMemoBinary -Arguments @('test-installation', '--output', 'json', '--runtime-namespace', $RuntimeNamespace) -ParseJson
}

function Get-KubeMemoInstallationStatus {
    [CmdletBinding()]
    param(
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    Invoke-KubeMemoBinary -Arguments @('status', '--output', 'json', '--runtime-namespace', $RuntimeNamespace) -ParseJson
}

function Open-KubeMemoTui {
    [CmdletBinding()]
    param(
        [switch]$IncludeRuntime,
        [switch]$DurableOnly,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [string[]]$Namespace,
        [int]$AutoRefreshSeconds = 3
    )

    $args = @('tui', '--runtime-namespace', $RuntimeNamespace, '--auto-refresh-seconds', $AutoRefreshSeconds.ToString())
    if (-not $DurableOnly) { $args += '--include-runtime' }
    if ($IncludeRuntime) { $args += '--include-runtime' }
    foreach ($ns in @($Namespace | Where-Object { $_ })) {
        $args += @('--namespace', $ns)
    }
    Invoke-KubeMemoBinary -Arguments $args -PassthruTerminal
}

function Get-KubeMemo {
    [CmdletBinding()]
    param(
        [switch]$IncludeRuntime,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [string[]]$Namespace
    )

    $args = @('get', '--output', 'json', '--runtime-namespace', $RuntimeNamespace)
    if ($IncludeRuntime) { $args += '--include-runtime' }
    foreach ($ns in @($Namespace | Where-Object { $_ })) {
        $args += @('--namespace', $ns)
    }
    Invoke-KubeMemoBinary -Arguments $args -ParseJson
}

function Find-KubeMemo {
    [CmdletBinding()]
    param(
        [string]$Text,
        [string]$NoteType,
        [string]$Kind,
        [string]$Namespace,
        [string]$Name,
        [switch]$IncludeRuntime,
        [switch]$DurableOnly,
        [switch]$IncludeExpired
    )

    $args = @('find', '--output', 'json')
    if ($Text) { $args += @('--text', $Text) }
    if ($NoteType) { $args += @('--note-type', $NoteType) }
    if ($Kind) { $args += @('--kind', $Kind) }
    if ($Namespace) { $args += @('--namespace', $Namespace) }
    if ($Name) { $args += @('--name', $Name) }
    if (-not $DurableOnly) { $args += '--include-runtime' }
    if ($IncludeRuntime) { $args += '--include-runtime' }
    if ($IncludeExpired) { $args += '--include-expired' }
    Invoke-KubeMemoBinary -Arguments $args -ParseJson
}

function Show-KubeMemo {
    [CmdletBinding()]
    param(
        [string]$Kind,
        [string]$Namespace,
        [string]$Name,
        [switch]$IncludeRuntime,
        [switch]$DurableOnly,
        [switch]$PassThru,
        [switch]$NoColor
    )

    $args = @('show')
    if ($Kind) { $args += @('--kind', $Kind) }
    if ($Namespace) { $args += @('--namespace', $Namespace) }
    if ($Name) { $args += @('--name', $Name) }
    if (-not $DurableOnly) { $args += '--include-runtime' }
    if ($IncludeRuntime) { $args += '--include-runtime' }
    if ($NoColor) { $args += '--no-color' }
    if ($PassThru) {
        return Invoke-KubeMemoBinary -Arguments $args
    }
    Invoke-KubeMemoBinary -Arguments $args -PassthruTerminal
}

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

    if ($PSCmdlet.ShouldProcess($Title, 'Create KubeMemo memo')) {
        $args = @('new', '--output', 'json', '--title', $Title, '--format', $Format, '--note-type', $NoteType, '--severity', $Severity, '--runtime-namespace', $RuntimeNamespace)
        if ($Summary) { $args += @('--summary', $Summary) }
        if ($Content) { $args += @('--content', $Content) }
        if ($OwnerTeam) { $args += @('--owner-team', $OwnerTeam) }
        if ($OwnerContact) { $args += @('--owner-contact', $OwnerContact) }
        foreach ($tag in @($Tags | Where-Object { $_ })) { $args += @('--tag', $tag) }
        if ($PSBoundParameters.ContainsKey('ExpiresAt')) { $args += @('--expires-at', $ExpiresAt.ToUniversalTime().ToString('o')) }
        if ($Temporary) { $args += '--temporary' }
        if ($AnnotateResource) { $args += '--annotate-resource' }
        if ($OutputPath) { $args += @('--output-path', $OutputPath) }
        if ($PSCmdlet.ParameterSetName -eq 'Resource') {
            $args += @('--kind', $Kind, '--name', $Name, '--api-version', $ApiVersion)
            if ($Namespace) { $args += @('--namespace', $Namespace) }
        } elseif ($PSCmdlet.ParameterSetName -eq 'Namespace') {
            $args += @('--target-namespace', $TargetNamespace)
        } else {
            $args += @('--app-name', $AppName)
            if ($AppInstance) { $args += @('--app-instance', $AppInstance) }
        }
        Invoke-KubeMemoBinary -Arguments $args -ParseJson
    }
}

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

    if ($PSCmdlet.ShouldProcess($Id, 'Update KubeMemo memo')) {
        $args = @('set', '--output', 'json', '--id', $Id, '--runtime-namespace', $RuntimeNamespace)
        if ($Title) { $args += @('--title', $Title) }
        if ($PSBoundParameters.ContainsKey('Summary')) { $args += @('--summary', $Summary) }
        if ($PSBoundParameters.ContainsKey('Content')) { $args += @('--content', $Content) }
        if ($PSBoundParameters.ContainsKey('Tags')) { foreach ($tag in @($Tags | Where-Object { $_ })) { $args += @('--tag', $tag) } }
        if ($PSBoundParameters.ContainsKey('ExpiresAt')) { $args += @('--expires-at', $ExpiresAt.ToUniversalTime().ToString('o')) }
        if ($Runtime) { $args += '--runtime' }
        if ($AnnotateResource) { $args += '--annotate-resource' }
        if ($OutputPath) { $args += @('--output-path', $OutputPath) }
        Invoke-KubeMemoBinary -Arguments $args -ParseJson
    }
}

function Remove-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [string]$Id,
        [switch]$ExpiredRuntimeOnly,
        [switch]$RemoveAnnotations,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    if ($PSCmdlet.ShouldProcess($(if ($Id) { $Id } else { 'runtime memos' }), 'Remove KubeMemo memo')) {
        $args = @('remove', '--output', 'json', '--runtime-namespace', $RuntimeNamespace)
        if ($Id) { $args += @('--id', $Id) }
        if ($ExpiredRuntimeOnly) { $args += '--expired-runtime-only' }
        if ($RemoveAnnotations) { $args += '--remove-annotations' }
        Invoke-KubeMemoBinary -Arguments $args -ParseJson
    }
}

function Export-KubeMemo {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Path,
        [switch]$IncludeRuntime
    )

    $args = @('export', '--output', 'json', '--path', $Path)
    if ($IncludeRuntime) { $args += '--include-runtime' }
    Invoke-KubeMemoBinary -Arguments $args -ParseJson
}

function Import-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory)]
        [string]$Path
    )

    if ($PSCmdlet.ShouldProcess($Path, 'Import KubeMemo manifests')) {
        Invoke-KubeMemoBinary -Arguments @('import', '--output', 'json', '--path', $Path) -ParseJson
    }
}

function Sync-KubeMemoGitOps {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [string]$Path = (Get-KubeMemoConfigInternal).GitOpsRepoPath,
        [ValidateSet('Export', 'Import')]
        [string]$Direction = 'Export'
    )

    if ($PSCmdlet.ShouldProcess($Path, "Sync KubeMemo GitOps ($Direction)")) {
        Invoke-KubeMemoBinary -Arguments @('sync-gitops', '--output', 'json', '--path', $Path, '--direction', $Direction.ToLowerInvariant()) -ParseJson
    }
}

function Test-KubeMemoGitOpsMode {
    [CmdletBinding()]
    param()

    Invoke-KubeMemoBinary -Arguments @('test-gitops-mode', '--output', 'json') -ParseJson
}

function Test-KubeMemoRuntimeStore {
    [CmdletBinding()]
    param(
        [string]$Namespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    Invoke-KubeMemoBinary -Arguments @('test-runtime-store', '--output', 'json', '--namespace', $Namespace) -ParseJson
}

function Clear-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'High')]
    param(
        [switch]$ExpiredOnly,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    if ($PSCmdlet.ShouldProcess($RuntimeNamespace, 'Clear KubeMemo runtime memos')) {
        $args = @('clear', '--output', 'json', '--runtime-namespace', $RuntimeNamespace)
        if ($ExpiredOnly) { $args += '--expired-only' }
        Invoke-KubeMemoBinary -Arguments $args -ParseJson
    }
}

function Get-KubeMemoActivity {
    [CmdletBinding()]
    param(
        [string]$Kind,
        [string]$Namespace,
        [string]$Name
    )

    $args = @('get-activity', '--output', 'json')
    if ($Kind) { $args += @('--kind', $Kind) }
    if ($Namespace) { $args += @('--namespace', $Namespace) }
    if ($Name) { $args += @('--name', $Name) }
    Invoke-KubeMemoBinary -Arguments $args -ParseJson
}

function Get-KubeMemoConfig {
    [CmdletBinding()]
    param()

    Invoke-KubeMemoBinary -Arguments @('get-config', '--output', 'json') -ParseJson
}
