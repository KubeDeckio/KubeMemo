function Install-KubeMemoManifest {
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter(Mandatory)]
        [string]$Manifest,

        [string]$Description = 'KubeMemo manifest'
    )

    if ($PSCmdlet.ShouldProcess($Description, 'Apply manifest')) {
        Invoke-KubeMemoKubectl -Arguments @('apply', '-f', '-') -InputObject $Manifest | Out-Null
    }
}
