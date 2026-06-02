<#
.SYNOPSIS
    Package a local release of pcc-toolkit.

.DESCRIPTION
    Builds the Go core binary and creates a portable release directory
    or archive containing the binary and install instructions.

    Release structure:
      release/
        pcc-toolkit-v0.2.0-windows-amd64/
          pcc-core.exe
          INSTALL.txt

.PARAMETER Version
    Version string for the release (default: reads from pyproject.toml).

.PARAMETER OutputDir
    Output directory for the release (default: release/ relative to repo root).

.PARAMETER Archive
    If set, creates a .zip archive of the release directory.

.PARAMETER OS
    Target GOOS (default: current OS).

.PARAMETER Arch
    Target GOARCH (default: current arch).

.EXAMPLE
    .\scripts\release.ps1
    .\scripts\release.ps1 -Version "0.2.0" -Archive
#>

param(
    [string]$Version,
    [string]$OutputDir,
    [switch]$Archive,
    [string]$OS,
    [string]$Arch
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path "$PSScriptRoot\.."

if (-not $OutputDir) {
    $OutputDir = Join-Path $repoRoot "release"
}

if (-not $Version) {
    $pyprojectPath = Join-Path $repoRoot "pyproject.toml"
    $versionLine = Get-Content -LiteralPath $pyprojectPath | Where-Object { $_ -match '^version\s*=\s*"(.+)"' } | Select-Object -First 1
    if (-not $versionLine) {
        throw "Could not read project version from $pyprojectPath"
    }
    $Version = $Matches[1]
}

if (-not $OS) {
    $OS = if ($IsWindows -or $env:OS -eq "Windows_NT") { "windows" } elseif ($IsLinux) { "linux" } else { "darwin" }
}
if (-not $Arch) {
    $Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
}

$releaseDirName = "pcc-toolkit-v${Version}-${OS}-${Arch}"
$releaseDir = Join-Path $OutputDir $releaseDirName

# Clean and recreate release directory
if (Test-Path -LiteralPath $releaseDir) {
    Remove-Item -LiteralPath $releaseDir -Recurse -Force
}
New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null

# Build the Go binary
Write-Host "=== Building pcc-core ==="
$buildScript = Join-Path $PSScriptRoot "build.ps1"
$buildOutputDir = Join-Path $repoRoot "build"
& $buildScript -Version $Version -OS $OS -Arch $Arch -OutputDir $buildOutputDir

# Copy binary to release directory
$exeSuffix = if ($OS -eq "windows") { ".exe" } else { "" }
$binaryName = "pcc-core${exeSuffix}"
$builtBinary = Join-Path $buildOutputDir $binaryName

if (-not (Test-Path -LiteralPath $builtBinary)) {
    throw "Binary not found at $builtBinary"
}

Copy-Item -LiteralPath $builtBinary -Destination $releaseDir

# Write INSTALL.txt
$installContent = @"
pcc-toolkit v$Version — $OS/$Arch
====================================

This is the Go core engine (pcc-core). It must be paired with the Python CLI.

Install steps:

1. Place pcc-core${exeSuffix} in your PATH, or in the build/ directory
   of the pcc-toolkit repository.

2. Install the Python CLI from the repository root:

   pip install .[cli]    # CLI

3. Verify installation:

   pcc-toolkit --version

4. (Optional) If you placed pcc-core in build/ instead of PATH:
   The tool locates pcc-core automatically in the repository build/ directory.

Files in this release:
  pcc-core${exeSuffix}    — Go core engine binary (all domain logic)
  INSTALL.txt             — This file

Published binaries (never committed to the repository):
  - pcc-core.exe (Windows)
  - pcc-core      (Linux/macOS)

These are built from core/cmd/pcc-core/ and excluded via .gitignore (*.exe, build/).
"@

$installPath = Join-Path $releaseDir "INSTALL.txt"
Set-Content -LiteralPath $installPath -Value $installContent -Encoding UTF8

Write-Host ""
Write-Host "=== Release prepared ==="
Write-Host "Directory: $releaseDir"
Write-Host "Contents:"
Get-ChildItem -LiteralPath $releaseDir | ForEach-Object { Write-Host "  $($_.Name)" }

if ($Archive) {
    $zipPath = Join-Path $OutputDir "${releaseDirName}.zip"
    if (Test-Path -LiteralPath $zipPath) {
        Remove-Item -LiteralPath $zipPath -Force
    }
    Compress-Archive -LiteralPath $releaseDir -DestinationPath $zipPath
    $zipSize = (Get-Item -LiteralPath $zipPath).Length
    $zipSizeKB = [math]::Round($zipSize / 1KB, 1)
    Write-Host ""
    Write-Host "Archive: $zipPath ($zipSizeKB KB)"
}
