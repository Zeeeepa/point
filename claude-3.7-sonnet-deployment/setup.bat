@echo off
setlocal enabledelayedexpansion

echo.
echo === Claude 3.7 Sonnet Deployment Setup ===
echo This script will help you set up the chatgpt-adapter for Claude 3.7 Sonnet via Cursor
echo.

REM Check if Docker is installed
where docker >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Docker is not installed. Please install Docker first.
    echo Visit https://docs.docker.com/get-docker/ for installation instructions.
    exit /b 1
)

REM Check if Docker Compose is installed
where docker-compose >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Docker Compose is not installed. Please install Docker Compose first.
    echo Visit https://docs.docker.com/compose/install/ for installation instructions.
    exit /b 1
)

REM Check if Docker is running
docker info >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Docker is not running. Please start Docker Desktop first.
    exit /b 1
)

REM Get Cursor session token
echo You need a valid Cursor session token (WorkosCursorSessionToken) to use Claude 3.7 Sonnet.
echo You can get this token by logging into Cursor (https://www.cursor.com) and extracting it from your browser cookies.
echo.
echo To get your token:
echo 1. Log in to Cursor (https://www.cursor.com)
echo 2. Open your browser's developer tools (F12 or right-click ^> Inspect)
echo 3. Go to the Application/Storage tab
echo 4. Find Cookies ^> https://www.cursor.com
echo 5. Copy the value of the 'WorkosCursorSessionToken' cookie
echo.

set /p CURSOR_TOKEN="Enter your Cursor session token (WorkosCursorSessionToken): "

if "!CURSOR_TOKEN!"=="" (
    echo No token provided. Setup cannot continue.
    exit /b 1
)

REM Create a token file
echo Saving token...
echo !CURSOR_TOKEN!> .cursor_token

REM Start the service
echo Starting chatgpt-adapter with Claude 3.7 Sonnet support...
docker-compose up -d

echo.
echo === Setup Complete! ===
echo The chatgpt-adapter is now running at http://localhost:8080
echo.
echo Available Claude 3.7 Models:
echo - claude-3.7-sonnet
echo - claude-3.7-sonnet-max
echo - claude-3.7-sonnet-thinking
echo - claude-3.7-sonnet-thinking-max
echo.
echo To test the deployment, run:
echo test_claude.bat
echo.
echo To stop the service:
echo docker-compose down
echo.

endlocal

