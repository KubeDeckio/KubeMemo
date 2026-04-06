function Get-KubeMemoActor {
    [CmdletBinding()]
    param()

    $whoAmIStatus = Invoke-KubeMemoKubectl -Arguments @('auth', 'whoami', '-o', 'json') -IgnoreErrors -PassStatus
    if ($whoAmIStatus.ExitCode -eq 0 -and -not [string]::IsNullOrWhiteSpace($whoAmIStatus.Text)) {
        try {
            $whoAmI = $whoAmIStatus.Text | ConvertFrom-Json
            $username = $whoAmI.status.userInfo.username
            if (-not [string]::IsNullOrWhiteSpace($username)) {
                return $username.Trim()
            }
        }
        catch {
        }
    }

    $status = Invoke-KubeMemoKubectl -Arguments @('config', 'view', '--minify', '-o', 'jsonpath={.contexts[0].context.user}') -IgnoreErrors -PassStatus
    if ($status.ExitCode -eq 0 -and -not [string]::IsNullOrWhiteSpace($status.Text)) {
        return $status.Text.Trim()
    }

    foreach ($candidate in @($env:KUBEMEMO_USER, $env:USER, $env:USERNAME)) {
        if (-not [string]::IsNullOrWhiteSpace($candidate)) {
            return $candidate.Trim()
        }
    }

    try {
        $identity = [System.Security.Principal.WindowsIdentity]::GetCurrent()
        if ($identity -and -not [string]::IsNullOrWhiteSpace($identity.Name)) {
            return $identity.Name.Trim()
        }
    }
    catch {
    }

    return 'unknown'
}
