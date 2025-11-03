@echo off
echo ========================================
echo ARI-S Test Suite Summary
echo ========================================
echo.

cd /d D:\Development\ARI-S

echo Running tests...
echo.

go test -v > test-output.txt 2>&1

echo Test Results:
echo ----------------------------------------
findstr /C:"PASS:" test-output.txt | find /C "PASS" > pass-count.txt
findstr /C:"FAIL:" test-output.txt | find /C "FAIL" > fail-count.txt

set /p PASS_COUNT=<pass-count.txt
set /p FAIL_COUNT=<fail-count.txt

echo Passed: %PASS_COUNT%
echo Failed: %FAIL_COUNT%
echo.

echo Failed Tests:
echo ----------------------------------------
findstr /C:"--- FAIL:" test-output.txt

echo.
echo ========================================
echo Full output saved to: test-output.txt
echo ========================================

del pass-count.txt fail-count.txt
pause
