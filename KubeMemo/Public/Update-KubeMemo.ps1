function Update-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [switch]$IncludeRbac,
        [string]$RuntimeNamespace = (Get-KubeMemoConfigInternal).RuntimeNamespace
    )

    foreach ($manifestName in 'DurableCrd', 'RuntimeCrd', 'RuntimeNamespace') {
        $manifest = Get-KubeMemoEmbeddedManifest -Name $manifestName
        if ($manifestName -eq 'RuntimeNamespace') {
            $manifest = $manifest -replace 'kubememo-runtime', $RuntimeNamespace
        }
        Install-KubeMemoManifest -Manifest $manifest -Description "KubeMemo $manifestName" -WhatIf:$WhatIfPreference -Confirm:$false
    }

    if ($IncludeRbac) {
        $manifest = (Get-KubeMemoEmbeddedManifest -Name Rbac) -replace 'kubememo-runtime', $RuntimeNamespace
        Install-KubeMemoManifest -Manifest $manifest -Description 'KubeMemo RBAC' -WhatIf:$WhatIfPreference -Confirm:$false
    }

    Get-KubeMemoInstallationStatus -RuntimeNamespace $RuntimeNamespace
}
