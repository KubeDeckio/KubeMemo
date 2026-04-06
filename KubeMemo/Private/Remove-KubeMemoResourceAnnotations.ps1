function Remove-KubeMemoResourceAnnotations {
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter(Mandatory)]
        [string]$Kind,

        [Parameter(Mandatory)]
        [string]$Name,

        [string]$Namespace
    )

    $payload = @{
        metadata = @{
            annotations = @{
                'notes.kubememo.io/enabled' = $null
                'notes.kubememo.io/has-note' = $null
                'notes.kubememo.io/summary' = $null
                'notes.kubememo.io/note-ref' = $null
                'notes.kubememo.io/runtime-enabled' = $null
            }
        }
    } | ConvertTo-Json -Compress -Depth 5

    $target = if ($Namespace) { "$Kind/$Name -n $Namespace" } else { "$Kind/$Name" }
    if ($PSCmdlet.ShouldProcess($target, 'Remove KubeMemo annotations from resource')) {
        $args = @('patch', $Kind, $Name, '--type', 'merge', '-p', $payload)
        if ($Namespace) {
            $args += @('-n', $Namespace)
        }
        Invoke-KubeMemoKubectl -Arguments $args | Out-Null
    }
}
