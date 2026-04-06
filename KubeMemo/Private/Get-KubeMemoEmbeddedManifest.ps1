function Get-KubeMemoEmbeddedManifest {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [ValidateSet('DurableCrd', 'RuntimeCrd', 'RuntimeNamespace', 'Rbac')]
        [string]$Name
    )

    $map = @{
        DurableCrd      = 'kubememonote.crd.yaml'
        RuntimeCrd      = 'kubememoruntimenote.crd.yaml'
        RuntimeNamespace = 'runtime-namespace.yaml'
        Rbac            = 'rbac.yaml'
    }

    $path = Join-Path -Path (Join-Path $PSScriptRoot '..\Assets') -ChildPath $map[$Name]
    Get-Content -Path $path -Raw
}
