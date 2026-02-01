function Test-DSCPullServerToken {
    [CmdletBinding()]
    param()
    if (-not $script:DSCPullServerSession.Token) {
        Write-Host "Aucun token n'est stocké."
        return $false
    }
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    try {
        $resp = Invoke-RestMethod -Uri "$($script:DSCPullServerSession.ServerUrl)/api/v1/my"xw -Headers $headers -Method GET -Verbose:$true
        Write-Host "Token valide pour l'utilisateur $($resp.email)"
        return $true
    } catch {
        Write-Host "Token invalide ou expiré"
        return $false
    }
}
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
    } elseif ($Credential) {
        $body = @{ username = $Credential.UserName; password = $Credential.GetNetworkCredential().Password } | ConvertTo-Json
        $resp = Invoke-RestMethod -Uri "$ServerUrl/api/v1/login" -Method POST -ContentType 'application/json' -Body $body
        if ($resp.token) {
            $script:DSCPullServerSession = @{ ServerUrl = $ServerUrl; Token = $resp.token; AuthType = 'Bearer' }
            return $true
        } else {
            throw "Échec de l'authentification."
        }
    } else {
        throw "Vous devez fournir soit -Token, soit -Credential."
    }
}

Export-ModuleMember -Function Connect-DSCPullServer
