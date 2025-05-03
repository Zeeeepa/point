@echo off
setlocal enabledelayedexpansion

echo ===================================================
echo Starting Web Claude OpenAI API compatible endpoint...
echo ===================================================
echo.

REM Call common installation script with Python
call "%~dp0install-common.bat" with-python
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install common dependencies.
    pause
    exit /b 1
)

REM Check if the ai-web-integration-agent directory exists
if not exist "%~dp0ai-web-integration-agent" (
    echo Error: ai-web-integration-agent directory not found.
    echo Please download the repository from https://github.com/Zeeeepa/point/tree/main/ai-web-integration-agent
    echo or clone it using: git clone https://github.com/Zeeeepa/point.git
    pause
    exit /b 1
)

REM Navigate to the ai-web-integration-agent directory
cd "%~dp0ai-web-integration-agent"

REM Create config directory if it doesn't exist
if not exist "config" mkdir config

REM Create logs directory if it doesn't exist
if not exist "logs" mkdir logs

REM Create screenshots directory if it doesn't exist
if not exist "screenshots" mkdir screenshots

REM Create browser_data directory if it doesn't exist
if not exist "browser_data" mkdir browser_data

REM Create a basic config file if it doesn't exist
if not exist "config\config.json" (
    echo Creating default configuration...
    (
        echo {
        echo   "claude_url": "https://claude.ai/",
        echo   "browser_user_data_dir": "browser_data",
        echo   "screenshot_dir": "screenshots",
        echo   "log_file": "logs/agent.log",
        echo   "headless": false,
        echo   "debug_mode": true,
        echo   "claude_login_required": true
        echo }
    ) > config\config.json
    echo Configuration file created at config\config.json
    echo NOTE: You will need to log in to Claude when the browser opens.
)

REM Install Python dependencies with specific versions
echo Installing Python dependencies...
python -m pip install --upgrade pip
python -m pip install selenium==4.15.2 webdriver-manager==4.0.1 pyautogui==0.9.54 requests==2.31.0 flask==3.0.0
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install Python dependencies.
    pause
    exit /b 1
)

REM Install Go dependencies
echo Installing Go dependencies...

REM Check if go.mod exists
if not exist "go.mod" (
    echo Initializing Go module...
    go mod init web-integration-agent
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to initialize Go module.
        pause
        exit /b 1
    )
)

REM Install Go dependencies with error checking
echo Installing chromedp dependency...
go get github.com/chromedp/chromedp
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install chromedp dependency.
    pause
    exit /b 1
)

echo Installing gorilla/mux dependency...
go get github.com/gorilla/mux
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install gorilla/mux dependency.
    pause
    exit /b 1
)

echo Installing rs/cors dependency...
go get github.com/rs/cors
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install rs/cors dependency.
    pause
    exit /b 1
)

REM Call Playwright installation script
call "%~dp0install-playwright.bat"
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install Playwright dependencies.
    pause
    exit /b 1
)

REM Install specific cdproto packages
echo Installing specific cdproto packages...
go get github.com/chromedp/cdproto/runtime/enable
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install cdproto/runtime/enable dependency.
    pause
    exit /b 1
)

go get github.com/tidwall/gjson@v1.18.0
if %ERRORLEVEL% NEQ 0 (
    echo Failed to install gjson dependency.
    pause
    exit /b 1
)

echo Running go mod tidy...
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo Failed to tidy go modules.
    pause
    exit /b 1
)

REM Build and run the web integration agent for Claude
echo Building and starting the Web Claude OpenAI API compatible endpoint...
go run web-integration-agent.go --mode claude --port 8081

REM If the program exits, pause to see any error messages
echo.
echo The Web Claude OpenAI API endpoint has stopped.
pause

