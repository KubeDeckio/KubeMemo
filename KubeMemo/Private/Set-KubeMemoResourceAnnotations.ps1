function Set-KubeMemoResourceAnnotations {
    [CmdletBinding(SupportsShouldProcess)]
    param(
        [Parameter(Mandatory)]
        [string]$Kind,

        [Parameter(Mandatory)]
        [string]$Name,

        [string]$Namespace,
        [string]$Summary,
        [string]$NoteRef,
        [switch]$RuntimeEnabled
    )

    $annotations = @{
        'notes.kubememo.io/enabled' = 'true'
        'notes.kubememo.io/has-note' = 'true'
    }

    if ($Summary) {
        $annotations['notes.kubememo.io/summary'] = $Summary
    }

    if ($NoteRef) {
        $annotations['notes.kubememo.io/note-ref'] = $NoteRef
    }

    if ($RuntimeEnabled) {
        $annotations['notes.kubememo.io/runtime-enabled'] = 'true'
    }

    $payload = @{
        metadata = @{
            annotations = $annotations
        }
    } | ConvertTo-Json -Compress -Depth 5

    $target = if ($Namespace) { "$Kind/$Name -n $Namespace" } else { "$Kind/$Name" }
    if ($PSCmdlet.ShouldProcess($target, 'Patch resource annotations for KubeMemo')) {
        $args = @('patch', $Kind, $Name, '--type', 'merge', '-p', $payload)
        if ($Namespace) {
            $args += @('-n', $Namespace)
        }
        Invoke-KubeMemoKubectl -Arguments $args | Out-Null
    }
}
