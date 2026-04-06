function Install-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [switch]$DurableOnly,
        [switch]$EnableRuntimeStore,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace,
        [switch]$GitOpsAware,
        [switch]$InstallRbac
    )

    $status = Test-KubeMemoInstallation -RuntimeNamespace $RuntimeNamespace
    if (-not $status.ClusterReachable) {
        throw 'Cluster connectivity verification failed.'
    }

    if (-not $status.DurableCrdInstalled) {
        Install-KubeMemoManifest -Manifest (Get-KubeMemoEmbeddedManifest -Name DurableCrd) -Description 'KubeMemo durable CRD' -WhatIf:$WhatIfPreference -Confirm:$false
    }

    $shouldEnableRuntime = $EnableRuntimeStore -or (-not $DurableOnly)
    if ($shouldEnableRuntime -and -not $status.RuntimeCrdInstalled) {
        Install-KubeMemoManifest -Manifest (Get-KubeMemoEmbeddedManifest -Name RuntimeCrd) -Description 'KubeMemo runtime CRD' -WhatIf:$WhatIfPreference -Confirm:$false
    }

    if ($shouldEnableRuntime -and -not (Test-KubeMemoNamespaceInstalled -Namespace $RuntimeNamespace)) {
        $manifest = (Get-KubeMemoEmbeddedManifest -Name RuntimeNamespace) -replace 'kubememo-runtime', $RuntimeNamespace
        Install-KubeMemoManifest -Manifest $manifest -Description "KubeMemo runtime namespace $RuntimeNamespace" -WhatIf:$WhatIfPreference -Confirm:$false
    }

    if ($InstallRbac) {
        $manifest = (Get-KubeMemoEmbeddedManifest -Name Rbac) -replace 'kubememo-runtime', $RuntimeNamespace
        Install-KubeMemoManifest -Manifest $manifest -Description 'KubeMemo RBAC' -WhatIf:$WhatIfPreference -Confirm:$false
    }

    $mode = Get-KubeMemoInstallationStatus -RuntimeNamespace $RuntimeNamespace
    if ($GitOpsAware -and $mode.Mode -eq 'GitOps with runtime store' -and -not $mode.RuntimeStoreSafe) {
        throw "Runtime store namespace '$RuntimeNamespace' is not considered safe for GitOps."
    }

    $mode
}
