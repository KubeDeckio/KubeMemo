@{
    RootModule        = 'KubeMemo.psm1'
    ModuleVersion     = '0.1.0'
    GUID              = '3fcdbe58-4906-4f57-91b7-f80995d8f53c'
    Author            = 'Richard Hooper'
    CompanyName       = 'Kubedeck'
    Copyright         = '(c) Richard Hooper. All rights reserved.'
    Description       = 'KubeMemo is a PowerShell-first Kubernetes operational memory tool.'
    PowerShellVersion = '7.2'
    FunctionsToExport = @(
        'Install-KubeMemo',
        'Uninstall-KubeMemo',
        'Update-KubeMemo',
        'Test-KubeMemoInstallation',
        'Get-KubeMemoInstallationStatus',
        'Open-KubeMemoTui',
        'Get-KubeMemo',
        'Find-KubeMemo',
        'Show-KubeMemo',
        'New-KubeMemo',
        'Set-KubeMemo',
        'Remove-KubeMemo',
        'Export-KubeMemo',
        'Import-KubeMemo',
        'Sync-KubeMemoGitOps',
        'Test-KubeMemoGitOpsMode',
        'Test-KubeMemoRuntimeStore',
        'Clear-KubeMemo',
        'Get-KubeMemoActivity',
        'Get-KubeMemoConfig'
    )
    CmdletsToExport   = @()
    VariablesToExport = @()
    AliasesToExport   = @()
    PrivateData       = @{
        PSData = @{
            Tags       = @('Kubernetes', 'PowerShell', 'Kubedeck', 'KubeMemo')
            LicenseUri = 'https://github.com/KubeDeckio/KubeMemo/blob/main/LICENSE'
            ProjectUri = 'https://github.com/kubedeck/KubeMemo'
            ReleaseNotes = 'https://github.com/KubeDeckio/KubeMemo/blob/main/README.md'
        }
    }
}
