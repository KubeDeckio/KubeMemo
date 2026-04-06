function Invoke-KubeMemoKubectl {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string[]]$Arguments,

        [string]$InputObject,

        [switch]$IgnoreErrors,

        [switch]$PassStatus
    )

    $command = Get-Command -Name kubectl -ErrorAction SilentlyContinue
    if (-not $command) {
        throw 'kubectl was not found on PATH.'
    }

    $joinedArgs = $Arguments -join ' '
    $result = if ($PSBoundParameters.ContainsKey('InputObject')) {
        $InputObject | & $command.Source @Arguments 2>&1
    }
    else {
        & $command.Source @Arguments 2>&1
    }

    $exitCode = $LASTEXITCODE
    if ($PassStatus) {
        return [pscustomobject]@{
            ExitCode = $exitCode
            Output = @($result)
            Text = (@($result) -join [Environment]::NewLine)
        }
    }

    if ($exitCode -ne 0 -and -not $IgnoreErrors) {
        throw "kubectl $joinedArgs failed."
    }

    return $result
}
