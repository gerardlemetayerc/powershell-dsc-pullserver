param(
    [string]$Output = "DSC-Admin-Console.exe",
    [string]$VersionFile = "build/version.json",
    [string]$CommitOverride = ""
)

$ErrorActionPreference = "Stop"

$isWindowsHost = ($env:OS -eq "Windows_NT")

function Write-Utf8NoBom {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Content
    )
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, $Content, $utf8NoBom)
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$versionFilePath = if ([System.IO.Path]::IsPathRooted($VersionFile)) { $VersionFile } else { Join-Path $repoRoot $VersionFile }

if (-not (Test-Path $versionFilePath)) {
    throw "Version file not found: $versionFilePath"
}

$cfg = Get-Content -Raw -Path $versionFilePath | ConvertFrom-Json
$version = [string]$cfg.version

if ([string]::IsNullOrWhiteSpace($version)) {
    throw "Version is empty in $versionFilePath"
}

# Keep raw version exactly as provided for product metadata and ldflags.
$versionRaw = $version.Trim()

# Extract numeric components for PE fixed file info.
$numericTokens = [regex]::Matches($versionRaw, '\d+') | ForEach-Object { [int]$_.Value }
if ($numericTokens.Count -eq 0) {
    $numericTokens = @(0, 0, 0, 0)
}
$major = if ($numericTokens.Count -ge 1) { $numericTokens[0] } else { 0 }
$minor = if ($numericTokens.Count -ge 2) { $numericTokens[1] } else { 0 }
$patch = if ($numericTokens.Count -ge 3) { $numericTokens[2] } else { 0 }
$build = if ($numericTokens.Count -ge 4) { $numericTokens[3] } else { 0 }
$fileVersion = "$major.$minor.$patch.$build"

if ([string]::IsNullOrWhiteSpace($CommitOverride)) {
    $commit = (git rev-parse --short=12 HEAD).Trim()
} else {
    $commit = $CommitOverride.Trim()
}

$buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
Push-Location $repoRoot

$targetGoos = if ([string]::IsNullOrWhiteSpace($env:GOOS)) {
    if ($Output.ToLower().EndsWith(".exe")) {
        "windows"
    } elseif ($isWindowsHost) {
        "windows"
    } else {
        "linux"
    }
} else {
    $env:GOOS
}
$targetGoarch = if ([string]::IsNullOrWhiteSpace($env:GOARCH)) { "amd64" } else { $env:GOARCH }
$targetWindows = ($targetGoos -eq "windows")

$tempDir = [System.IO.Path]::GetTempPath()
$versionInfoPath = Join-Path $tempDir "versioninfo.generated.json"
$sysoPath = Join-Path $repoRoot ("resource_windows_{0}.syso" -f $targetGoarch)

$versionInfo = @{
    FixedFileInfo = @{
        FileVersion    = @{ Major = $major; Minor = $minor; Patch = $patch; Build = $build }
        ProductVersion = @{ Major = $major; Minor = $minor; Patch = $patch; Build = $build }
        FileFlagsMask  = "3f"
        FileFlags      = "00"
        FileOS         = "040004"
        FileType       = "01"
        FileSubType    = "00"
    }
    StringFileInfo = @{
        Comments         = ""
        CompanyName      = [string]$cfg.companyName
        FileDescription  = [string]$cfg.fileDescription
        FileVersion      = $versionRaw
        InternalName     = [string]$cfg.internalName
        LegalCopyright   = [string]$cfg.legalCopyright
        OriginalFilename = [string]$cfg.originalFilename
        ProductName      = [string]$cfg.productName
        ProductVersion   = $versionRaw
    }
    VarFileInfo = @{
        Translation = @{
            LangID    = "0409"
            CharsetID = "04B0"
        }
    }
}

$versionInfoJson = $versionInfo | ConvertTo-Json -Depth 8
Write-Utf8NoBom -Path $versionInfoPath -Content $versionInfoJson

try {
    if ($targetWindows) {
        go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

        $goBin = (go env GOPATH).Trim()
        $goversioninfoSuffix = if ($isWindowsHost) { ".exe" } else { "" }
        $goversioninfo = Join-Path $goBin ("bin/goversioninfo{0}" -f $goversioninfoSuffix)
        if (-not (Test-Path $goversioninfo)) {
            throw "goversioninfo not found at $goversioninfo"
        }

        & $goversioninfo -64 -o $sysoPath $versionInfoPath
    }

    $ldflags = "-X go-dsc-pull/internal/buildinfo.Version=$versionRaw -X go-dsc-pull/internal/buildinfo.Commit=$commit -X go-dsc-pull/internal/buildinfo.BuildDate=$buildDate"
    go build -ldflags $ldflags -o $Output
}
finally {
    if (Test-Path $sysoPath) {
        Remove-Item $sysoPath -Force
    }
    if (Test-Path $versionInfoPath) {
        Remove-Item $versionInfoPath -Force
    }
    Pop-Location
}

Write-Host "Built $Output"
Write-Host "  ProductVersion: $versionRaw"
Write-Host "  FileVersion:    $fileVersion"
Write-Host "  Commit:         $commit"
Write-Host "  BuildDate(UTC): $buildDate"
