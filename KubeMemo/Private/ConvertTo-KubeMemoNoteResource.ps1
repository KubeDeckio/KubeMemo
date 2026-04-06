function ConvertTo-KubeMemoNoteResource {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$ResourceName,

        [Parameter(Mandatory)]
        [string]$Namespace,

        [Parameter(Mandatory)]
        [hashtable]$Spec
    )

    $labels = @{
        'notes.kubememo.io/type' = $Spec.noteType
    }
    if ($Spec.target.appRef.name) {
        $labels['app.kubernetes.io/name'] = $Spec.target.appRef.name
    }

    @{
        apiVersion = 'notes.kubememo.io/v1alpha1'
        kind = 'Memo'
        metadata = @{
            name = $ResourceName
            namespace = $Namespace
            labels = $labels
        }
        spec = $Spec
        status = @{
            state = 'active'
            expired = $false
        }
    }
}
