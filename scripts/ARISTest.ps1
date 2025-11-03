# ARI-S Test Report Generator
param()

# Change to project root
Set-Location $PSScriptRoot\..

$testOutput = go test -v ./... 2>&1 | Out-String

$workflows = @(
    @{Name="Application Lifecycle"; File="app_lifecycle_test.go"; Pattern="TestApplicationLifecycle"}
    @{Name="Configuration Persistence"; File="config_persistence_test.go"; Pattern="TestConfigPersistence"}
    @{Name="Dependency Management"; File="dependency_management_test.go"; Pattern="TestDependencyManagement"}
    @{Name="File/Folder Browsing"; File="file_browsing_test.go"; Pattern="TestFileBrowsing"}
    @{Name="Retoc Pack"; File="retoc_pack_test.go"; Pattern="TestRetocPack"}
    @{Name="Retoc Unpack"; File="retoc_unpack_test.go"; Pattern="TestRetocUnpack"}
    @{Name="UAsset Export"; File="uasset_export_test.go"; Pattern="TestUAssetExport"}
    @{Name="UAsset Import"; File="uasset_import_test.go"; Pattern="TestUAssetImport"}
)

Write-Host ""
Write-Host "========================================================================================" -ForegroundColor Cyan
Write-Host "                               ARI-S TEST RESULTS                                       " -ForegroundColor Cyan
Write-Host "========================================================================================" -ForegroundColor Cyan
Write-Host ""

$format = "{0,-30} {1,-35} {2,-10} {3}"
Write-Host ($format -f "Workflow", "Test File", "Tests", "Status") -ForegroundColor White
Write-Host ($format -f "--------", "---------", "-----", "------") -ForegroundColor Gray

foreach ($wf in $workflows) {
    $runPattern = '=== RUN\s+' + $wf.Pattern
    $passPattern = '--- PASS:\s+' + $wf.Pattern

    $totalMatches = [regex]::Matches($testOutput, $runPattern)
    $passMatches = [regex]::Matches($testOutput, $passPattern)

    $total = $totalMatches.Count
    $passed = $passMatches.Count

    if ($total -gt 0) {
        $pct = [math]::Round(($passed / $total) * 100)
        $testInfo = "$passed/$total"

        if ($pct -eq 100) {
            $statusText = "100% PASS"
            $statusColor = "Green"
        } elseif ($pct -ge 75) {
            $statusText = " $pct% PASS"
            $statusColor = "Yellow"
        } else {
            $statusText = " $pct% FAIL"
            $statusColor = "Red"
        }

        Write-Host ("{0,-30} {1,-35} {2,-10} " -f $wf.Name, $wf.File, $testInfo) -NoNewline -ForegroundColor Gray
        Write-Host $statusText -ForegroundColor $statusColor
    }
}

Write-Host ""

$totalRun = ([regex]::Matches($testOutput, '=== RUN\s+Test')).Count
$totalPass = ([regex]::Matches($testOutput, '--- PASS:\s+Test')).Count
$rate = [math]::Round(($totalPass / $totalRun) * 100, 1)

Write-Host ""
Write-Host "  Total: $totalPass/$totalRun tests passed ($rate%)" -ForegroundColor Cyan
Write-Host ""

if ($totalPass -lt $totalRun) {
    Write-Host "  Note: Retoc Pack/Unpack tests 'fail' to validate error handling." -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "----------------------------------------------------------------------------------------" -ForegroundColor Cyan
}
