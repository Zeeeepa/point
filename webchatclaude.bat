@echo off
echo Starting Web Claude OpenAI API compatible endpoint...
echo.

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Go is not installed. Installing Go...
    echo Please wait, this might take a few minutes...
    powershell -Command "& {Invoke-WebRequest -Uri 'https://go.dev/dl/go1.21.0.windows-amd64.msi' -OutFile 'go_installer.msi'; Start-Process -FilePath 'msiexec.exe' -ArgumentList '/i', 'go_installer.msi', '/quiet' -Wait; Remove-Item 'go_installer.msi'}"
    echo Go has been installed.
    echo.
)

REM Check if Python is installed
where python >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Python is not installed. Installing Python...
    echo Please wait, this might take a few minutes...
    powershell -Command "& {Invoke-WebRequest -Uri 'https://www.python.org/ftp/python/3.10.0/python-3.10.0-amd64.exe' -OutFile 'python_installer.exe'; Start-Process -FilePath 'python_installer.exe' -ArgumentList '/quiet', 'InstallAllUsers=1', 'PrependPath=1' -Wait; Remove-Item 'python_installer.exe'}"
    echo Python has been installed.
    echo.
)

REM Check if the ai-web-integration-agent directory exists
if not exist "%~dp0ai-web-integration-agent" (
    echo Error: ai-web-integration-agent directory not found.
    echo Please make sure the directory exists in the same location as this batch file.
    pause
    exit /b 1
)

REM Navigate to the ai-web-integration-agent directory
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
)

REM Install Python dependencies
echo Installing Python dependencies...
pip install selenium webdriver-manager pyautogui requests flask

REM Install Go dependencies
echo Installing Go dependencies...
go mod init web-integration-agent
go get github.com/chromedp/chromedp
go get github.com/gorilla/mux
go get github.com/rs/cors
go get github.com/playwright-community/playwright-go

REM Install playwright browsers
echo Installing playwright browsers...
go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install playwright browsers.
    pause
    exit /b 1
)

REM Build and run the web integration agent for Claude
echo Building and starting the Web Claude OpenAI API compatible endpoint...
go run web-integration-agent.go --mode claude --port 8081

REM If the program exits, pause to see any error messages
pause
