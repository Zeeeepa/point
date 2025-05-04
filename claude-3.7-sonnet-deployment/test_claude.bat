@echo off
setlocal enabledelayedexpansion

echo.
echo === Testing Claude 3.7 Sonnet Deployment ===
echo.

REM Check if curl is installed
where curl >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo curl is not installed. Please install curl first.
    exit /b 1
)

REM Check if the service is running
curl -s http://localhost:8080/v1/models >nul
if %ERRORLEVEL% neq 0 (
    echo The chatgpt-adapter service is not running. Please start it first.
    echo Run: docker-compose up -d
    exit /b 1
)

echo Service is running. Testing Claude 3.7 Sonnet...

REM Get token from file
if exist .cursor_token (
    set /p TOKEN=<.cursor_token
) else (
    echo No token file found. Using the API without authentication.
    set TOKEN=
)

REM Set headers
if not "!TOKEN!"=="" (
    set AUTH_HEADER=Authorization: Bearer !TOKEN!
) else (
    set AUTH_HEADER=
)

REM Make the API call
echo Sending test request to Claude 3.7 Sonnet...
echo.

curl -s -X POST http://localhost:8080/v1/chat/completions ^
    -H "Content-Type: application/json" ^
    -H "!AUTH_HEADER!" ^
    -d "{\"model\":\"cursor/claude-3.7-sonnet\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello, are you Claude 3.7 Sonnet? Please confirm your model name and provide a brief greeting.\"}]}"

echo.
echo.
echo If you received a proper response above, Claude 3.7 Sonnet is working correctly.
echo If you received an error, please check your token and Docker setup.
echo.

endlocal

