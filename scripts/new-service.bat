@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

title New Service Generator

:: Colors
set GREEN=[92m
set RED=[91m
set CYAN=[96m
set YELLOW=[93m
set NC=[0m

echo.
echo %CYAN%========================================%NC%
echo %CYAN%  New Service Generator%NC%
echo %CYAN%  Based on service-boilerplate%NC%
echo %CYAN%========================================%NC%
echo.

:: Check arguments
if "%~1"=="" goto :show_help
if /i "%~1"=="help" goto :show_help
if /i "%~1"=="/?" goto :show_help

set PROJECT_NAME=%~1
set OUTPUT_DIR=%~2

if "%OUTPUT_DIR%"=="" set OUTPUT_DIR=.

:: Validate project name
echo %CYAN%==^> Creating new project: %PROJECT_NAME%%NC%

:: Check if directory exists
if exist "%OUTPUT_DIR%\%PROJECT_NAME%" (
    echo %RED%[ERROR] Directory %OUTPUT_DIR%\%PROJECT_NAME% already exists%NC%
    exit /b 1
)

:: Create directory
mkdir "%OUTPUT_DIR%\%PROJECT_NAME%"
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Failed to create directory%NC%
    exit /b 1
)

echo %GREEN%[OK] Directory created%NC%

:: Clone boilerplate
echo %CYAN%==^> Cloning boilerplate...%NC%
git clone --depth 1 https://github.com/warlocknt/service-boilerplate.git "%OUTPUT_DIR%\%PROJECT_NAME%\temp"
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Failed to clone boilerplate%NC%
    rmdir /s /q "%OUTPUT_DIR%\%PROJECT_NAME%" 2>nul
    exit /b 1
)

:: Move files from temp
cd "%OUTPUT_DIR%\%PROJECT_NAME%\temp"
for /f %%a in ('dir /b /a') do (
    move "%%a" "..\" >nul 2>&1
)
cd ..
rmdir /s /q temp

:: Remove .git and initialize new
echo %CYAN%==^> Initializing new git repository...%NC%
if exist .git rmdir /s /q .git
git init
git add .
git commit -m "Initial commit based on service-boilerplate"

echo %GREEN%[OK] Git repository initialized%NC%

:: Update go.mod module name
echo %CYAN%==^> Updating module name...%NC%
go mod edit -module %PROJECT_NAME%
if %errorlevel% neq 0 (
    echo %YELLOW%[WARNING] Failed to update go.mod, doing manually...%NC%
    powershell -Command "(Get-Content go.mod) -replace 'module service-boilerplate', 'module %PROJECT_NAME%' | Set-Content go.mod"
)
echo %GREEN%[OK] Module name updated to: %PROJECT_NAME%%NC%

:: Update config
echo %CYAN%==^> Updating configuration...%NC%
powershell -Command "(Get-Content configs/config.yaml) -replace 'service-boilerplate', '%PROJECT_NAME%' | Set-Content configs/config.yaml"
echo %GREEN%[OK] Configuration updated%NC%

:: Update README
echo %CYAN%==^> Updating README...%NC%
powershell -Command "(Get-Content README.md) -replace 'service-boilerplate', '%PROJECT_NAME%' | Set-Content README.md"
echo %GREEN%[OK] README updated%NC%

:: Clean up
echo %CYAN%==^> Cleaning up...%NC%
if exist service.exe del service.exe
if exist build rmdir /s /q build
if exist logs rmdir /s /q logs
if exist coverage.out del coverage.out
if exist coverage.html del coverage.html
echo %GREEN%[OK] Cleanup complete%NC%

:: Create new commit with changes
git add .
git commit -m "chore: rename module to %PROJECT_NAME%"

echo.
echo %GREEN%========================================%NC%
echo %GREEN%  Project created successfully!%NC%
echo %GREEN%========================================%NC%
echo.
echo Location: %OUTPUT_DIR%\%PROJECT_NAME%
echo.
echo Next steps:
echo   cd %PROJECT_NAME%
echo   scripts\build.bat
echo.
echo Or:
echo   go mod tidy
echo   go build -o %PROJECT_NAME%.exe ./cmd/service-boilerplate
echo.
exit /b 0

:show_help
echo.
echo New Service Generator
echo =====================
echo.
echo Usage: new-service.bat ^<project-name^> [output-directory]
echo.
echo Examples:
echo   new-service.bat my-api-service
echo   new-service.bat payment-service C:\Projects
echo   new-service.bat notification-service ..
echo.
echo This will:
echo   1. Clone service-boilerplate from GitHub
echo   2. Rename Go module
echo   3. Update configuration files
echo   4. Initialize new git repository
echo   5. Clean up build artifacts
echo.
exit /b 0
