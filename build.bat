@echo off
setlocal

:: ==========================================
:: 1. Environment Setup (MSYS2 / MinGW)
:: ==========================================
set "MSYS2_ROOT=D:\Scoop\apps\msys2\current"
set "MINGW64_BIN=%MSYS2_ROOT%\mingw64\bin"

if not exist "%MSYS2_ROOT%" (
    echo MSYS2_ROOT not found: %MSYS2_ROOT%
    echo Please install MSYS2 or update MSYS2_ROOT in this script.
    exit /b 1
)

if not exist "%MINGW64_BIN%" (
    echo MINGW64 bin not found: %MINGW64_BIN%
    exit /b 1
)

:: Clear potential conflicting global variables
set "C_INCLUDE_PATH="
set "CPLUS_INCLUDE_PATH="
set "LIBRARY_PATH="
set "GCC_EXEC_PREFIX="

:: Ensure MINGW64_BIN is at front of PATH
echo %PATH% | find /I "%MINGW64_BIN%" >nul
if errorlevel 1 (
    set "PATH=%MINGW64_BIN%;%PATH%"
)

:: ==========================================
:: 2. Versioning & Preparation
:: ==========================================
echo Getting git version...
for /f "tokens=*" %%g in ('git rev-parse --short HEAD') do (set GIT_VERSION=%%g)
if not defined GIT_VERSION (
    echo WARNING: Could not get git version. Using 'dev'.
    set GIT_VERSION=dev
)
echo Git Version: %GIT_VERSION%

:: Create build directory if it doesn't exist
if not exist "build" (
    echo Creating 'build' directory...
    mkdir "build"
)

:: Set Ldflags
set LDFLAGS=-ldflags="-X main.Version=%GIT_VERSION% -s -w"

:: ==========================================
:: 3. Build for Windows (amd64)
:: ==========================================
echo.
echo [1/6] Building shipyard for Windows (amd64)...

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
set CC=gcc
set CXX=g++



echo.
echo [2/6] Building shipyard-server for Windows (amd64)...

go build %LDFLAGS% -o build/shipyard-server-windows-amd64.exe ./cmd/shipyard-server

if %errorlevel% neq 0 (
    echo [ERROR] shipyard-server Windows build FAILED!
    exit /b %errorlevel%
) else (
    echo [SUCCESS] Created build/shipyard-server-windows-amd64.exe
)

echo.
echo [3/6] Building shipyard-cli for Windows (amd64)...

go build %LDFLAGS% -o build/shipyard-cli-windows-amd64.exe ./cmd/shipyard-cli

if %errorlevel% neq 0 (
    echo [ERROR] shipyard-cli Windows build FAILED!
    exit /b %errorlevel%
) else (
    echo [SUCCESS] Created build/shipyard-cli-windows-amd64.exe
)

:: ==========================================
:: 4. Build for Linux (amd64)
:: ==========================================
echo.
echo [4/6] Building shipyard for Linux (amd64)...

set GOOS=linux
set GOARCH=amd64
:: NOTE: Cross-compiling CGO from Windows to Linux requires a specific 
:: cross-compiler (not just MinGW). We use CGO_ENABLED=0 for safety here.
:: If your app requires CGO (e.g. SQLite), this step requires Docker or WSL.
set CGO_ENABLED=0



echo.
echo [5/6] Building shipyard-server for Linux (amd64)...

go build %LDFLAGS% -o build/shipyard-server-linux-amd64 ./cmd/shipyard-server

if %errorlevel% neq 0 (
    echo [ERROR] shipyard-server Linux build FAILED!
    exit /b %errorlevel%
) else (
    echo [SUCCESS] Created build/shipyard-server-linux-amd64
)

echo.
echo [6/6] Building shipyard-cli for Linux (amd64)...

go build %LDFLAGS% -o build/shipyard-cli-linux-amd64 ./cmd/shipyard-cli

if %errorlevel% neq 0 (
    echo [ERROR] shipyard-cli Linux build FAILED!
    exit /b %errorlevel%
) else (
    echo [SUCCESS] Created build/shipyard-cli-linux-amd64
)

echo.
echo ==========================================
echo All builds finished successfully.
echo Artifacts are in the 'build' folder.
echo ==========================================

endlocal