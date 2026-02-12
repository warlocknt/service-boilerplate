@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

title Service Boilerplate Build

:: Colors (ANSI escape codes)
set GREEN=[92m
set RED=[91m
set CYAN=[96m
set YELLOW=[93m
set NC=[0m

:: Configuration
set BINARY_NAME=service-boilerplate
set BUILD_DIR=.\build
set TEST_TIMEOUT=120s
set COVERAGE_FILE=coverage.out
set HTML_REPORT=coverage.html

:: Disable CGO by default (required for cross-platform builds)
set CGO_ENABLED=0

echo.
echo %CYAN%========================================%NC%
echo %CYAN%  Service Boilerplate Build Script%NC%
echo %CYAN%========================================%NC%
echo.

:: Check if Go is installed
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Go is not installed or not in PATH%NC%
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

:: Setup VS environment for race detector support
call :setup_vs
if defined VSCMD_VER (
    echo %GREEN%[OK] Visual Studio 2022 environment loaded%NC%
) else (
    echo %YELLOW%[WARNING] Visual Studio not found. Some features disabled.%NC%
)

:: Parse command
if "%~1"=="" goto :do_all
if /i "%~1"=="all" goto :do_all
if /i "%~1"=="test" goto :do_test
if /i "%~1"=="test-fast" goto :do_test_fast
if /i "%~1"=="build" goto :do_build_with_tests
if /i "%~1"=="build-only" goto :do_build_only
if /i "%~1"=="build-all" goto :do_build_all
if /i "%~1"=="build-win" goto :do_build_win
if /i "%~1"=="build-linux" goto :do_build_linux
if /i "%~1"=="coverage" goto :do_coverage
if /i "%~1"=="check" goto :do_check
if /i "%~1"=="clean" goto :do_clean
if /i "%~1"=="deps" goto :do_deps
if /i "%~1"=="ci" goto :do_ci
if /i "%~1"=="help" goto :do_help
if /i "%~1"=="/?" goto :do_help
if /i "%~1"=="-h" goto :do_help
if /i "%~1"=="--help" goto :do_help

echo %RED%[ERROR] Unknown command: %~1%NC%
goto :do_help

:: ============================================
:: Main Commands
:: ============================================

:do_all
call :run_tests
if %errorlevel% neq 0 exit /b 1
call :clean_build
call :build_binary "%BINARY_NAME%.exe"
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  Build Completed Successfully!%NC%
echo %GREEN%========================================%NC%
echo.
echo Binary location: %BUILD_DIR%\%BINARY_NAME%.exe
exit /b 0

:do_test
call :run_tests
pause
exit /b %errorlevel%

:do_test_fast
call :run_tests_fast
pause
exit /b %errorlevel%

:do_build_with_tests
call :run_tests
if %errorlevel% neq 0 exit /b 1
call :clean_build
call :build_binary "%BINARY_NAME%.exe"
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  Build Completed Successfully!%NC%
echo %GREEN%========================================%NC%
pause
exit /b 0

:do_build_only
call :clean_build
call :build_binary "%BINARY_NAME%.exe"
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  Build Completed Successfully!%NC%
echo %GREEN%========================================%NC%
pause
exit /b 0

:do_build_all
call :run_tests
if %errorlevel% neq 0 exit /b 1
call :clean_build
call :build_all_platforms
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  Multi-Platform Build Complete!%NC%
echo %GREEN%========================================%NC%
pause
exit /b 0

:do_build_win
call :clean_build
call :build_windows
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  Windows Build Complete!%NC%
echo %GREEN%========================================%NC%
pause
exit /b 0

:do_build_linux
call :clean_build
call :build_linux
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  Linux Build Complete!%NC%
echo %GREEN%========================================%NC%
pause
exit /b 0

:do_coverage
call :generate_coverage
pause
exit /b 0

:do_check
call :run_checks
pause
exit /b 0

:do_clean
call :clean_build
pause
exit /b 0

:do_deps
call :download_deps
pause
exit /b 0

:do_ci
call :run_ci
pause
exit /b 0

:do_help
echo.
echo Service Boilerplate Build Script
echo =================================
echo.
echo Usage: build.bat [command]
echo.
echo Commands:
echo   all         Run tests and build (default)
echo   test        Run all tests
echo   test-fast   Run all tests (alias for test)
echo   build       Build binary (only if tests pass)
echo   build-only  Build without running tests
echo   build-all   Cross-compile: Windows (.exe) + Linux (amd64, arm64)
echo   build-win   Build for current platform (Windows .exe)
echo   build-linux Cross-compile for Linux (amd64, arm64)
echo   coverage    Generate coverage report
echo   check       Run formatting and vet
echo   clean       Clean build artifacts
echo   deps        Download dependencies
echo   ci          Full CI pipeline
echo   help        Show this help message
echo.
echo Examples:
echo   build.bat              - Run tests and build for current platform
echo   build.bat test         - Run tests only
echo   build.bat build-only   - Build without tests
echo   build.bat build-all    - Build for Windows + Linux
echo   build.bat build-linux  - Build only Linux binaries (from Windows)
echo   build.bat ci           - Full CI pipeline
echo.
echo Notes:
echo   - Race detector (-race) is NOT available on Windows due to MSVC/CGO issues
echo   - Use Linux/Mac or WSL for race detection: make test
echo.
exit /b 0

:: ============================================
:: Subroutines
:: ============================================

:setup_vs
if defined VSCMD_VER exit /b 0

:: Try to find vcvarsall.bat via vswhere
set "VSWHERE=%ProgramFiles(x86)%\Microsoft Visual Studio\Installer\vswhere.exe"
if exist "%VSWHERE%" (
    for /f "delims=" %%i in ('"%VSWHERE%" -latest -property installationPath') do (
        if exist "%%i\VC\Auxiliary\Build\vcvarsall.bat" (
            call "%%i\VC\Auxiliary\Build\vcvarsall.bat" x64 >nul 2>&1
            if defined VSCMD_VER exit /b 0
        )
    )
)

:: Fallback: check common paths
set "VS2022_COMMUNITY=C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Auxiliary\Build\vcvarsall.bat"
set "VS2022_PRO=C:\Program Files\Microsoft Visual Studio\2022\Professional\VC\Auxiliary\Build\vcvarsall.bat"
set "VS2022_ENT=C:\Program Files\Microsoft Visual Studio\2022\Enterprise\VC\Auxiliary\Build\vcvarsall.bat"
set "VS2022_BUILD=C:\Program Files (x86)\Microsoft Visual Studio\2022\BuildTools\VC\Auxiliary\Build\vcvarsall.bat"

if exist "%VS2022_COMMUNITY%" (
    call "%VS2022_COMMUNITY%" x64 >nul 2>&1
    if defined VSCMD_VER exit /b 0
)
if exist "%VS2022_PRO%" (
    call "%VS2022_PRO%" x64 >nul 2>&1
    if defined VSCMD_VER exit /b 0
)
if exist "%VS2022_ENT%" (
    call "%VS2022_ENT%" x64 >nul 2>&1
    if defined VSCMD_VER exit /b 0
)
if exist "%VS2022_BUILD%" (
    call "%VS2022_BUILD%" x64 >nul 2>&1
    if defined VSCMD_VER exit /b 0
)

exit /b 0

:clean_build
echo.
echo %CYAN%==^> Cleaning build directory...%NC%
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
if exist %COVERAGE_FILE% del /f /q %COVERAGE_FILE%
if exist %HTML_REPORT% del /f /q %HTML_REPORT%
mkdir %BUILD_DIR%
echo %GREEN%[OK] Clean complete%NC%
exit /b 0

:download_deps
echo.
echo %CYAN%==^> Downloading dependencies...%NC%
go mod download
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Failed to download dependencies%NC%
    exit /b 1
)
go mod tidy
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Failed to tidy dependencies%NC%
    exit /b 1
)
echo %GREEN%[OK] Dependencies ready%NC%
exit /b 0

:run_tests
echo.
echo %CYAN%==^> Running tests...%NC%
:: Note: Race detector disabled on Windows due to MSVC/CGO compatibility issues
:: Use Linux/Mac or WSL for race detection: make test
go test -v -timeout %TEST_TIMEOUT% ./...
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Tests failed!%NC%
    exit /b 1
)
echo %GREEN%[OK] All tests passed%NC%
exit /b 0

:run_tests_fast
echo.
echo %CYAN%==^> Running tests (fast mode)...%NC%
go test -v -timeout 60s ./...
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Tests failed!%NC%
    exit /b 1
)
echo %GREEN%[OK] All tests passed%NC%
exit /b 0

:generate_coverage
echo.
echo %CYAN%==^> Generating coverage report...%NC%
go test -coverprofile=%COVERAGE_FILE% -timeout %TEST_TIMEOUT% ./...
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Tests failed!%NC%
    exit /b 1
)
go tool cover -func=%COVERAGE_FILE%
echo %GREEN%[OK] Coverage report generated: %COVERAGE_FILE%%NC%
go tool cover -html=%COVERAGE_FILE% -o %HTML_REPORT%
echo %GREEN%[OK] HTML report generated: %HTML_REPORT%%NC%
exit /b 0

:build_binary
echo.
echo %CYAN%==^> Building binary: %~1...%NC%
set OUTPUT=%BUILD_DIR%\%~1
go build -ldflags="-s -w" -o %OUTPUT% .\cmd\service-boilerplate
if %errorlevel% neq 0 (
    echo %RED%[ERROR] Build failed!%NC%
    exit /b 1
)
echo %GREEN%[OK] Build complete: %OUTPUT%%NC%
exit /b 0

:build_windows
set GOOS=windows
set GOARCH=amd64
call :build_binary "%BINARY_NAME%-windows-amd64.exe"
exit /b 0

:build_linux
set GOOS=linux
set GOARCH=amd64
call :build_binary "%BINARY_NAME%-linux-amd64"
set GOARCH=arm64
call :build_binary "%BINARY_NAME%-linux-arm64"
exit /b 0

:build_all_platforms
echo.
echo %CYAN%==^> Building for multiple platforms...%NC%
call :build_windows
call :build_linux
echo %GREEN%[OK] Multi-platform build complete%NC%
exit /b 0

:run_checks
echo.
echo %CYAN%==^> Running code checks...%NC%
echo Formatting code...
go fmt ./...
echo Running go vet...
go vet ./...
echo %GREEN%[OK] Checks passed%NC%
exit /b 0

:run_ci
echo %YELLOW%[WARNING] Running in CI mode...%NC%
call :download_deps
call :run_checks
call :run_tests
call :clean_build
call :build_all_platforms
echo.
echo %GREEN%========================================%NC%
echo %GREEN%  CI Pipeline Completed Successfully!%NC%
echo %GREEN%========================================%NC%
exit /b 0
