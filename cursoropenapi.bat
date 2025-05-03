@echo off
setlocal enabledelayedexpansion

echo ===================================================
echo Starting Cursor OpenAI API compatible endpoint...
echo ===================================================
echo.

REM Call common installation script
call "%~dp0install-common.bat"
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install common dependencies.
    pause
    exit /b 1
)

REM Check if the chatgpt-adapter-main directory exists
if not exist "%~dp0chatgpt-adapter-main" (
    echo Error: chatgpt-adapter-main directory not found.
    echo Please download the repository from https://github.com/Zeeeepa/point/tree/main/chatgpt-adapter-main
    echo or clone it using: git clone https://github.com/Zeeeepa/point.git
    pause
    exit /b 1
)

REM Navigate to the chatgpt-adapter-main directory
cd "%~dp0chatgpt-adapter-main"

REM Check if dependencies are installed
if not exist "go.sum" (
    echo Installing dependencies...
    go mod download
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to download dependencies.
        pause
        exit /b 1
    )
)

REM Call Playwright installation script
call "%~dp0install-playwright.bat"
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install Playwright dependencies.
    pause
    exit /b 1
)

REM Fix missing go.sum entries
echo Fixing go.sum entries...
go get github.com/tidwall/gjson@v1.18.0
if %ERRORLEVEL% NEQ 0 (
    echo Failed to get gjson dependency.
    pause
    exit /b 1
)

go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo Failed to fix go.sum entries.
    pause
    exit /b 1
)

REM Create logs directory if it doesn't exist
if not exist "logs" mkdir logs

REM Build and run the adapter
echo Building and starting the Cursor OpenAI API compatible endpoint...
go run main.go --mode cursor

REM If the program exits, pause to see any error messages
echo.
echo The Cursor OpenAI API endpoint has stopped.
pause

