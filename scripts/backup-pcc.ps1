# backup-pcc.ps1
# Creates backups of original ME2 OT game files before modding.
# Backups are saved to ~\Desktop\MassEffect2Backups

param(
    [string]$GameDir = "C:\Program Files\EA Games\Mass Effect 2\BioGame\CookedPC",
    [string]$BackupDir = "$env:USERPROFILE\Desktop\MassEffect2Backups"
)

$ErrorActionPreference = "Stop"
New-Item -ItemType Directory -Path $BackupDir -Force | Out-Null

# Files currently being modded — edit this list as needed.
$files = @(
    "BioD_CitHub_400LowerWing_LOC_INT.pcc",
    "BIOGame_INT.tlk"
)

foreach ($f in $files) {
    $src = Join-Path $GameDir $f
    $dst = Join-Path $BackupDir $f

    if (-not (Test-Path $src)) {
        Write-Warning "Not found: $src"
        continue
    }

    if (Test-Path $dst) {
        Write-Host "Already backed up: $f"
    } else {
        Copy-Item $src $dst -Force
        Write-Host "Backed up: $f              ($([math]::Round((Get-Item $src).Length / 1MB, 1)) MB)"
    }
}

Write-Host "Done. Backups in $BackupDir"
