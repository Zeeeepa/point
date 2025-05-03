@echo off
REM Playwright installation script
REM This script handles installation of Playwright and its browser dependencies

echo Checking Playwright installation...

REM Check if Playwright is already installed
go run github.com/playwright-community/playwright-go/cmd/playwright --version >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo Playwright is already installed.
) else (
    echo Installing Playwright Go dependency...
    go get github.com/playwright-community/playwright-go@v0.4.0
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to install playwright-go dependency.
        exit /b 1
    )
    echo Playwright Go dependency installed successfully.
)

REM Install browser binaries
echo Installing Playwright browsers...
go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install Playwright browsers.
    exit /b 1
)

REM Verify browser installation
echo Verifying Playwright browser installation...
go run github.com/playwright-community/playwright-go/cmd/playwright test-browser
if %ERRORLEVEL% NEQ 0 (
    echo Playwright browser verification failed.
    exit /b 1
)

echo Playwright and browser dependencies installed successfully.
exit /b 0

