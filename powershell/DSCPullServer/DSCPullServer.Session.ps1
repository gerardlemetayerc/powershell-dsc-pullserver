# Stocke la session courante dans une variable de module
$script:DSCPullServerSession = @{}

function Connect-DSCPullServer {
    [CmdletBinding()]
    param(
        [string]$ServerUrl = 'http://localhost:8080',
        [string]$Token,
        [System.Management.Automation.PSCredential]$Credential
    )
    if ($Token) {
        $script:DSCPullServerSession = @{ ServerUrl = $ServerUrl; Token = $Token; AuthType = 'Token' }
        return $true
    } elseif ($Credential) {
        $body = @{ username = $Credential.UserName; password = $Credential.GetNetworkCredential().Password } | ConvertTo-Json
        $resp = Invoke-RestMethod -Uri "$ServerUrl/api/v1/login" -Method POST -ContentType 'application/json' -Body $body
        if ($resp.token) {
            $script:DSCPullServerSession = @{ ServerUrl = $ServerUrl; Token = $resp.token; AuthType = 'Bearer' }
            Write-Host $resp
            return $true
        } else {
            throw "Échec de l'authentification."
        }
    } else {
        throw "Vous devez fournir soit -Token, soit -Credential."
    }
}

Export-ModuleMember -Function Connect-DSCPullServer
