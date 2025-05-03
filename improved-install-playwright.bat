@echo off
REM Playwright installation script
REM This script handles installation of Playwright and its browser dependencies
setlocal enabledelayedexpansion

REM Check if verbose mode is enabled
set VERBOSE=0
if "%~1"=="verbose" set VERBOSE=1

REM Create logs directory if it doesn't exist
if not exist "%~dp0logs" mkdir "%~dp0logs"

REM Get the current date and time for the log filename
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "LOGDATE=%dt:~0,8%-%dt:~8,6%"
set "LOGFILE=%~dp0logs\install-playwright_%LOGDATE%.log"

REM Start logging
echo Starting Playwright installation script at %TIME% > "%LOGFILE%"
echo Environment information: >> "%LOGFILE%"
echo OS Version: >> "%LOGFILE%"
ver >> "%LOGFILE%"
echo System PATH: >> "%LOGFILE%"
echo %PATH% >> "%LOGFILE%"
echo. >> "%LOGFILE%"
echo ===================================================== >> "%LOGFILE%"
echo. >> "%LOGFILE%"

REM Tee function to display and log output
set "TEE=powershell -Command "$input | Tee-Object -FilePath '%LOGFILE%' -Append""

if %VERBOSE% EQU 1 (
    echo Verbose mode enabled. Logging to %LOGFILE% | %TEE%
)

echo Checking Playwright installation... | %TEE%

REM Check if Playwright is already installed
echo Checking if Playwright is already installed... | %TEE%
go run github.com/playwright-community/playwright-go/cmd/playwright --version >nul 2>&1
set EXIT_CODE=%ERRORLEVEL%
echo Playwright version check completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% EQU 0 (
    echo Playwright is already installed. | %TEE%
) else (
    echo Installing Playwright Go dependency... | %TEE%
    echo Running: go get github.com/playwright-community/playwright-go@v0.4.0 | %TEE%
    go get github.com/playwright-community/playwright-go@v0.4.0
    set EXIT_CODE=%ERRORLEVEL%
    echo Playwright Go dependency installation completed with exit code %EXIT_CODE% | %TEE%
    
    if %EXIT_CODE% NEQ 0 (
        echo Failed to install playwright-go dependency. | %TEE%
        echo Error details: | %TEE%
        go get github.com/playwright-community/playwright-go@v0.4.0 2>&1 | %TEE%
        exit /b 1
    )
    echo Playwright Go dependency installed successfully. | %TEE%
)

REM Install browser binaries
echo Installing Playwright browsers... | %TEE%
echo Running: go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps | %TEE%
go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps
set EXIT_CODE=%ERRORLEVEL%
echo Playwright browsers installation completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install Playwright browsers. | %TEE%
    echo Error details: | %TEE%
    go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps 2>&1 | %TEE%
    exit /b 1
)

REM Verify browser installation
echo Verifying Playwright browser installation... | %TEE%
echo Running: go run github.com/playwright-community/playwright-go/cmd/playwright test-browser | %TEE%
go run github.com/playwright-community/playwright-go/cmd/playwright test-browser
set EXIT_CODE=%ERRORLEVEL%
echo Playwright browser verification completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Playwright browser verification failed. | %TEE%
    echo Error details: | %TEE%
    go run github.com/playwright-community/playwright-go/cmd/playwright test-browser 2>&1 | %TEE%
    exit /b 1
)

echo Playwright and browser dependencies installed successfully. | %TEE%
exit /b 0

endlocal

