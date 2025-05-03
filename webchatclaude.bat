@echo off
echo Starting Web Claude OpenAI API compatible endpoint...
echo.

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Go is not installed. Would you like to install it? (Y/N)
    set /p install_choice=
    if /i "%install_choice%"=="Y" (
        echo Installing Go...
        echo Please wait, this might take a few minutes...
        powershell -Command "& {Invoke-WebRequest -Uri 'https://go.dev/dl/go1.21.0.windows-amd64.msi' -OutFile 'go_installer.msi'; Start-Process -FilePath 'msiexec.exe' -ArgumentList '/i', 'go_installer.msi', '/quiet' -Wait; Remove-Item 'go_installer.msi'}"
        echo Go has been installed.
        echo.
    ) else (
        echo Go installation skipped. This script requires Go to run.
        pause
        exit /b 1
    )
)

REM Check if Python is installed
where python >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Python is not installed. Would you like to install it? (Y/N)
    set /p install_choice=
    if /i "%install_choice%"=="Y" (
        echo Installing Python...
        echo Please wait, this might take a few minutes...
        powershell -Command "& {Invoke-WebRequest -Uri 'https://www.python.org/ftp/python/3.10.0/python-3.10.0-amd64.exe' -OutFile 'python_installer.exe'; Start-Process -FilePath 'python_installer.exe' -ArgumentList '/quiet', 'InstallAllUsers=1', 'PrependPath=1' -Wait; Remove-Item 'python_installer.exe'}"
        echo Python has been installed.
        echo.
    ) else (
        echo Python installation skipped. This script requires Python to run.
        pause
        exit /b 1
    )
)

REM Navigate to the ai-web-integration-agent directory
echo Checking for ai-web-integration-agent directory...
if not exist "%~dp0ai-web-integration-agent" (
    echo Error: ai-web-integration-agent directory not found.
    echo Please ensure the directory exists in the same location as this batch file.
    pause
    exit /b 1
)
cd "%~dp0ai-web-integration-agent"

REM Create config directory if it doesn't exist
if not exist "config" mkdir config

REM Create a basic config file if it doesn't exist
if not exist "config\config.json" (
    echo Creating default configuration...
    echo { > config\config.json
    echo   "claude_url": "https://claude.ai/", >> config\config.json
    echo   "browser_user_data_dir": "browser_data", >> config\config.json
    echo   "screenshot_dir": "screenshots", >> config\config.json
    echo   "log_file": "logs/agent.log", >> config\config.json
    echo   "headless": false, >> config\config.json
    echo   "debug_mode": true, >> config\config.json
    echo   "claude_login_required": true >> config\config.json
    echo } >> config\config.json
    
    REM Create logs directory for log file
    if not exist "logs" mkdir logs
    
    echo Configuration file created at config\config.json
    echo Note: This configuration stores browser data in the browser_data directory,
    echo which may contain sensitive information. Please secure this directory appropriately.
)

REM Install Python dependencies
echo Installing Python dependencies...
pip install selenium webdriver-manager pyautogui requests flask

REM Install Go dependencies
echo Installing Go dependencies...
if not exist "go.mod" (
    echo Initializing Go module...
    go mod init web-integration-agent
)
go get github.com/chromedp/chromedp
go get github.com/gorilla/mux
go get github.com/rs/cors

REM Check if port 8081 is already in use
powershell -Command "& {if (Test-NetConnection -ComputerName localhost -Port 8081 -InformationLevel Quiet -WarningAction SilentlyContinue -ErrorAction SilentlyContinue) { Write-Host 'Warning: Port 8081 is already in use. The service may fail to start.' }}"

REM Build and run the web integration agent for Claude
echo Building and starting the Web Claude OpenAI API compatible endpoint...
echo.
echo Note: This will create an OpenAI API compatible endpoint at http://localhost:8081
echo You can use this endpoint with any OpenAI API client by setting the base URL to http://localhost:8081
echo.
go run web-integration-agent.go --mode claude --port 8081

REM If the program exits, pause to see any error messages
pause
