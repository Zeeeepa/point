@echo off
setlocal enabledelayedexpansion

REM Create logs directory if it doesn't exist
if not exist "%~dp0logs" mkdir "%~dp0logs"

REM Get the current date and time for the log filename
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "LOGDATE=%dt:~0,8%-%dt:~8,6%"
set "LOGFILE=%~dp0logs\cursoropenapi_%LOGDATE%.log"

echo ===================================================
echo Starting Cursor OpenAI API compatible endpoint...
echo Log file: %LOGFILE%
echo ===================================================
echo.
echo Press any key to start execution...
pause > nul

REM Start logging
echo Starting Cursor OpenAI API endpoint at %TIME% > "%LOGFILE%"
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

REM Check if the chatgpt-adapter-main directory exists
if not exist "%~dp0chatgpt-adapter-main" (
    echo Error: chatgpt-adapter-main directory not found. | %TEE%
    echo Please download the repository from https://github.com/Zeeeepa/point/tree/main/chatgpt-adapter-main | %TEE%
    echo or clone it using: git clone https://github.com/Zeeeepa/point.git | %TEE%
    echo. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Call common installation script with verbose output
echo Calling install-common.bat... | %TEE%
call "%~dp0install-common.bat" verbose
set EXIT_CODE=%ERRORLEVEL%
echo install-common.bat completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install common dependencies. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Navigate to the chatgpt-adapter-main directory
echo Changing directory to chatgpt-adapter-main... | %TEE%
cd "%~dp0chatgpt-adapter-main"
if %ERRORLEVEL% NEQ 0 (
    echo Failed to change directory to chatgpt-adapter-main. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Check if dependencies are installed
if not exist "go.sum" (
    echo Installing dependencies... | %TEE%
    go mod download
    set EXIT_CODE=%ERRORLEVEL%
    echo go mod download completed with exit code %EXIT_CODE% | %TEE%
    
    if %EXIT_CODE% NEQ 0 (
        echo Failed to download dependencies. | %TEE%
        echo Press any key to exit... | %TEE%
        pause > nul
        exit /b 1
    )
)

REM Call Playwright installation script with verbose output
echo Calling install-playwright.bat... | %TEE%
call "%~dp0install-playwright.bat" verbose
set EXIT_CODE=%ERRORLEVEL%
echo install-playwright.bat completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install Playwright dependencies. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Fix missing go.sum entries
echo Fixing go.sum entries... | %TEE%
echo Running: go get github.com/tidwall/gjson@v1.18.0 | %TEE%
go get github.com/tidwall/gjson@v1.18.0
set EXIT_CODE=%ERRORLEVEL%
echo go get github.com/tidwall/gjson completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to get gjson dependency. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

echo Running: go mod tidy | %TEE%
go mod tidy
set EXIT_CODE=%ERRORLEVEL%
echo go mod tidy completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to fix go.sum entries. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Create logs directory if it doesn't exist
if not exist "logs" (
    echo Creating logs directory... | %TEE%
    mkdir logs
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to create logs directory. | %TEE%
        echo Press any key to exit... | %TEE%
        pause > nul
        exit /b 1
    )
)

REM Build and run the adapter
echo Building and starting the Cursor OpenAI API compatible endpoint... | %TEE%
echo Running: go run main.go --mode cursor | %TEE%
go run main.go --mode cursor
set EXIT_CODE=%ERRORLEVEL%
echo Cursor OpenAI API endpoint exited with code %EXIT_CODE% | %TEE%

REM If the program exits, pause to see any error messages
echo. | %TEE%
echo The Cursor OpenAI API endpoint has stopped. | %TEE%
echo Press any key to exit... | %TEE%
pause > nul

endlocal

