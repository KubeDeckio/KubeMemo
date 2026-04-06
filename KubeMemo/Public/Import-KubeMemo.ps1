function Import-KubeMemo {
    [CmdletBinding(SupportsShouldProcess, ConfirmImpact = 'Medium')]
    param(
        [Parameter(Mandatory)]
        [string]$Path
    )

    foreach ($file in Get-ChildItem -Path $Path -File | Where-Object Extension -in '.json', '.yaml', '.yml') {
        if ($PSCmdlet.ShouldProcess($file.FullName, 'Apply KubeMemo manifest')) {
            Invoke-KubeMemoKubectl -Arguments @('apply', '-f', '-') -InputObject (Get-Content -Path $file.FullName -Raw) | Out-Null
        }
    }
}
