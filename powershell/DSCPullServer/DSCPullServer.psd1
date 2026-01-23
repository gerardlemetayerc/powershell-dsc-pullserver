@{
    # Module manifest for module 'DSCPullServer'
    RootModule = 'DSCPullServer.psm1'
    ModuleVersion = '0.1.0'
    GUID = 'b1e1e1e1-0000-4000-8000-000000000001'
    Author = 'Charles GERARD-LE METAYER'
    Copyright = '(c) 2026 Charles GERARD-LE METAYER. Tous droits réservés.'
    Description = 'Module PowerShell pour interagir avec le DSC Pull Server Go.'
    PowerShellVersion = '5.1'
    FunctionsToExport = @('Get-DSCPullServerAgent','Get-DSCPullServerReport','Connect-DSCPullServer','Add-DSCPullServerModule','Get-DSCPullServerModule','Remove-DSCPullServerModule','Get-DSCPullServerConfiguration','Add-DSCPullServerConfiguration','Remove-DSCPullServerConfiguration')
    CmdletsToExport = @()
    VariablesToExport = @()
    AliasesToExport = @()
    PrivateData = @{}
}