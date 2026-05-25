<#
.SYNOPSIS
    Build the pcc-core Go binary.

.DESCRIPTION
    Standalone build script for the pcc-core Go engine. No Python required.
    Produces a single binary in build/ with ldflags version injection.

.PARAMETER Version
    Version string to inject (default: reads from core go.mod or falls back to "0.0.0-dev").

.PARAMETER OutputDir
    Output directory for the binary (default: build/ relative to repo root).

.PARAMETER OS
    Target GOOS (default: current OS). Use "windows", "linux", "darwin".

.PARAMETER Arch
    Target GOARCH (default: current arch). Use "amd64", "arm64".

.EXAMPLE
    .\scripts\build.ps1
    .\scripts\build.ps1 -Version "0.3.0"
    .\scripts\build.ps1 -OS linux -Arch amd64
#>

param(
    [string]$Version,
    [string]$OutputDir,
    [string]$OS,
    [string]$Arch
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path "$PSScriptRoot\.."
$coreMain = Join-Path $repoRoot "core\cmd\pcc-core"

if (-not $OutputDir) {
    $OutputDir = Join-Path $repoRoot "build"
}

if (-not $Version) {
    $Version = "0.0.0-dev"
}

# Platform defaults
if (-not $OS) {
    $OS = if ($IsWindows -or $env:OS -eq "Windows_NT") { "windows" } elseif ($IsLinux) { "linux" } else { "darwin" }
}
if (-not $Arch) {
    $Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
}

$exeSuffix = if ($OS -eq "windows") { ".exe" } else { "" }
$binaryName = "pcc-core${exeSuffix}"
$outputPath = Join-Path $OutputDir $binaryName

$env:GOOS = $OS
$env:GOARCH = $Arch

$ldflags = "-s -w -X main.version=$Version"

New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

Write-Host "Building pcc-core $Version for ${OS}/${Arch} -> $outputPath"

Push-Location -LiteralPath $coreMain
try {
    $buildArgs = @(
        "build",
        "-ldflags", $ldflags,
        "-o", $outputPath,
        "."
    )
    & go @buildArgs
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

$size = (Get-Item -LiteralPath $outputPath).Length
$sizeKB = [math]::Round($size / 1KB, 1)
Write-Host "Built: $outputPath ($sizeKB KB)"
