@echo off
echo Cleaning up ARI.S processes...

echo Killing Node.js processes...
taskkill /f /im node.exe 2>nul

echo Killing ARI.S processes...
taskkill /f /im ARI-S.exe 2>nul

echo Checking for processes on port 9245...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :9245') do (
    echo Killing process %%a on port 9245...
    taskkill /f /pid %%a 2>nul
)

echo Cleanup complete!
pause
