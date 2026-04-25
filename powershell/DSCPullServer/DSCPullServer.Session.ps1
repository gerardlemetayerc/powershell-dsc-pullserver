function Test-DSCPullServerToken {
    [CmdletBinding()]
    param()
    if (-not $script:DSCPullServerSession.Token) {
        Write-Host "Aucun token n'est stocké."
        return $false
    }
    $headers = @{ Authorization = "$($script:DSCPullServerSession.AuthType) $($script:DSCPullServerSession.Token)" }
    try {
        $resp = Invoke-RestMethod -Uri "$($script:DSCPullServerSession.ServerUrl)/api/v1/my" -Headers $headers -Method GET -Verbose:$true
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
        $invokeParams = @{
            Uri             = "$ServerUrl/api/v1/login"
            Method          = 'POST'
            ContentType     = 'application/json'
            Body            = $body
            SessionVariable = 'webSession'
        }
        if ($PSVersionTable.PSVersion.Major -lt 6) {
            $invokeParams['UseBasicParsing'] = $true
        }
        $resp = Invoke-WebRequest @invokeParams
        $jwtToken = $null

        if ($webSession -and $webSession.Cookies) {
            try {
                $cookie = $webSession.Cookies.GetCookies($ServerUrl) | Where-Object { $_.Name -eq 'jwt_token' } | Select-Object -First 1
                if ($cookie) {
                    $jwtToken = $cookie.Value
                }
            } catch {
                # Fallback handled below via Set-Cookie header parsing.
            }
        }

        if (-not $jwtToken) {
            $setCookie = $resp.Headers['Set-Cookie']
            if ($setCookie -match 'jwt_token=([^;]+)') {
                $jwtToken = $matches[1]
            }
        }

        if ($jwtToken) {
            $script:DSCPullServerSession = @{ ServerUrl = $ServerUrl; Token = $jwtToken; AuthType = 'Bearer' }
            return $true
        }

        throw "Echec de l'authentification: aucun jeton JWT recu."
    } else {
        throw "Vous devez fournir soit -Token, soit -Credential."
    }
}

Export-ModuleMember -Function Connect-DSCPullServer
