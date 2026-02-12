function Add-DSCPullServerModule {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Path
    )
        if (-not $script:DSCPullServerSession.Token) {
        throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
        $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/modules/upload"
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    $file = Get-Item $Path
    $form = @{ file = $file }
    Invoke-RestMethod -Uri $uri -Method Post -Headers $headers -Form $form
}

function Get-DSCPullServerModule {
    [CmdletBinding()]
    param()
    if (-not $script:DSCPullServerSession.Token) {
           throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
        $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/modules"
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    Invoke-RestMethod -Uri $uri -Method Get -Headers $headers
}

function Remove-DSCPullServerModule {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$ModuleName
    )
    if (-not $script:DSCPullServerSession.Token) {
            throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
        $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/modules/delete?name=$ModuleName"
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    Invoke-RestMethod -Uri $uri -Method Delete -Headers $headers
}

# Fonctions pour gérer les configuration_models DSC
function Get-DSCPullServerConfiguration {
    [CmdletBinding()]
    param()
    if (-not $script:DSCPullServerSession.Token) {
        throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
    $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/configuration_models"
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    Invoke-RestMethod -Uri $uri -Method Get -Headers $headers
}

function Add-DSCPullServerConfiguration {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$Path
    )
    if (-not $script:DSCPullServerSession.Token) {
        throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
    $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/configuration_models/upload"
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    if ($PSVersionTable.PSVersion.Major -ge 7) {
        $file = Get-Item $Path
        $form = @{ file = $file }
        Invoke-RestMethod -Uri $uri -Method Post -Headers $headers -Form $form
    } else {
        # Compatible Windows PowerShell 5.1 : upload multipart
        Add-Type -AssemblyName System.Net.Http
        $client = New-Object System.Net.Http.HttpClient
        $multipart = New-Object System.Net.Http.MultipartFormDataContent
        $fileStream = [System.IO.File]::OpenRead($Path)
        $fileName = [System.IO.Path]::GetFileName($Path)
        $fileContent = New-Object System.Net.Http.StreamContent($fileStream)
        $fileContent.Headers.Add('Content-Disposition', "form-data; name=\"file\"; filename=\"$fileName\"")
        $multipart.Add($fileContent, 'file', $fileName)
        foreach ($k in $headers.Keys) { $client.DefaultRequestHeaders.Add($k, $headers[$k]) }
        $response = $client.PostAsync($uri, $multipart).Result
        $fileStream.Close()
        if (-not $response.IsSuccessStatusCode) {
            $body = $response.Content.ReadAsStringAsync().Result
            throw "Erreur HTTP $($response.StatusCode): $body"
        }
        $response.Content.ReadAsStringAsync().Result
    }
}

function Remove-DSCPullServerConfiguration {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$ConfigurationName
    )
    if (-not $script:DSCPullServerSession.Token) {
        throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
    $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/configuration_models/delete?name=$ConfigurationName"
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    Invoke-RestMethod -Uri $uri -Method Delete -Headers $headers
}

# Importe la gestion de session
. "$PSScriptRoot/DSCPullServer.Session.ps1"

function Get-DSCPullServerAgent {
    [CmdletBinding()]
    param(
        [string]$NodeName,
        [bool]$HasErrorLastReport
    )
    if (-not $script:DSCPullServerSession.Token) {
            throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
    $params = @{}
    if ($NodeName) { $params['node_name'] = $NodeName }
    if ($HasErrorLastReport -eq $true) { $params['has_error_last_report'] = 'true' }
    elseif ($HasErrorLastReport -eq $false) { $params['has_error_last_report'] = 'false' }
    $queryString = ($params.GetEnumerator() | ForEach-Object {"$($_.Key)=$($_.Value)"}) -join "&"
    if ($queryString) {
        $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/agents?$queryString"
    } else {
        $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/agents"
    }
    $authType = $script:DSCPullServerSession.AuthType
    Invoke-RestMethod -Uri $uri -Method GET -Headers @{ Authorization = "$authType $($script:DSCPullServerSession.Token)" }
}

function Get-DSCPullServerReport {
    [CmdletBinding()]
    param(
        [string]$AgentId
    )
    if (-not $AgentId) { throw "AgentId requis" }
    if (-not $script:DSCPullServerSession.Token) {
            throw "Vous devez d'abord appeler Connect-DSCPullServer."
    }
        $uri = "$($script:DSCPullServerSession.ServerUrl)/api/v1/agents/$AgentId/reports"
    $authType = $script:DSCPullServerSession.AuthType
    Invoke-RestMethod -Uri $uri -Method GET -Headers @{ Authorization = "$authType $($script:DSCPullServerSession.Token)" }
}

    Export-ModuleMember -Function Get-DSCPullServerAgent,Get-DSCPullServerReport,Connect-DSCPullServer,Add-DSCPullServerModule,Get-DSCPullServerModule,Remove-DSCPullServerModule,Get-DSCPullServerConfiguration,Add-DSCPullServerConfiguration,Remove-DSCPullServerConfiguration
