# restore.ps1
# Restores original game files from backup.
# Run backup-pcc.ps1 first if you haven't already.

param(
    [string]$GameDir = "C:\Program Files\EA Games\Mass Effect 2\BioGame\CookedPC",
    [string]$BackupDir = "$env:USERPROFILE\Desktop\MassEffect2Backups"
)

$ErrorActionPreference = "Stop"

$files = @(
    "BioD_CitHub_400LowerWing_LOC_INT.pcc",
    "BIOGame_INT.tlk"
)

foreach ($f in $files) {
    $gamePath = Join-Path $GameDir $f
    $backupPath = Join-Path $BackupDir $f

    if (-not (Test-Path $backupPath)) {
        Write-Error "Backup not found: $backupPath. Run backup-pcc.ps1 first."
        exit 1
    }
    Copy-Item $backupPath $gamePath -Force
    Write-Host "Restored: $f"
}

Write-Host "Done."
