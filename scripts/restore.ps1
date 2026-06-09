# restore.ps1
# Restores original game files from backup AND copies modified files from output/.
# Run backup-pcc.ps1 first if you haven't already.

param(
    [string]$GameDir = "C:\Program Files\EA Games\Mass Effect 2\BioGame\CookedPC",
    [string]$BackupDir = "$env:USERPROFILE\Desktop\MassEffect2Backups",
    [string]$OutputDir = "$PSScriptRoot\..\output",
    [switch]$RestoreOnly
)

$ErrorActionPreference = "Stop"

# ------------------------------------------------------------------
# Files managed by this script.
#   Name     = original filename in the game folder
#   ModName  = modified filename in output/ (null if not yet modded)
# ------------------------------------------------------------------
$files = @(
    @{Name = "BioD_CitHub_400LowerWing_LOC_INT.pcc"; ModName = "BioD_CitHub_400LowerWing_POC.pcc"},
    @{Name = "BIOGame_INT.tlk";                      ModName = $null}
)

foreach ($entry in $files) {
    $f = $entry.Name
    $gamePath = Join-Path $GameDir $f
    $backupPath = Join-Path $BackupDir $f

    if ($RestoreOnly -or ($null -eq $entry.ModName)) {
        if (-not (Test-Path $backupPath)) {
            Write-Error "Backup not found: $backupPath. Run backup-pcc.ps1 first."
            exit 1
        }
        Copy-Item $backupPath $gamePath -Force
        Write-Host "Restored original: $f"
    } else {
        $moddedPath = Join-Path $OutputDir $entry.ModName
        if (Test-Path $moddedPath) {
            Copy-Item $moddedPath $gamePath -Force
            Write-Host "Installed modded:  $f  <-  $($entry.ModName)"
        } else {
            Write-Warning "Modded file not found: $moddedPath (restoring original instead)"
            Copy-Item $backupPath $gamePath -Force
        }
    }
}

Write-Host "Done."
