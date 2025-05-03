@echo off
setlocal enabledelayedexpansion

REM Create logs directory if it doesn't exist
if not exist "%~dp0logs" mkdir "%~dp0logs"

REM Get the current date and time for the log filename
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "LOGDATE=%dt:~0,8%-%dt:~8,6%"
set "LOGFILE=%~dp0logs\webchatcopilot_%LOGDATE%.log"

echo ===================================================
echo Starting Web Copilot OpenAI API compatible endpoint...
echo Log file: %LOGFILE%
echo ===================================================
echo.
echo Press any key to start execution...
pause > nul

REM Start logging
echo Starting Web Copilot OpenAI API endpoint at %TIME% > "%LOGFILE%"
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

REM Check if the ai-web-integration-agent directory exists
if not exist "%~dp0ai-web-integration-agent" (
    echo Error: ai-web-integration-agent directory not found. | %TEE%
    echo Please download the repository from https://github.com/Zeeeepa/point/tree/main/ai-web-integration-agent | %TEE%
    echo or clone it using: git clone https://github.com/Zeeeepa/point.git | %TEE%
    echo. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Call common installation script with Python and verbose output
echo Calling install-common.bat with-python verbose... | %TEE%
call "%~dp0install-common.bat" with-python verbose
set EXIT_CODE=%ERRORLEVEL%
echo install-common.bat completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install common dependencies. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Navigate to the ai-web-integration-agent directory
echo Changing directory to ai-web-integration-agent... | %TEE%
cd "%~dp0ai-web-integration-agent"
if %ERRORLEVEL% NEQ 0 (
    echo Failed to change directory to ai-web-integration-agent. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Create required directories
echo Creating required directories... | %TEE%
for %%D in (config logs screenshots browser_data) do (
    if not exist "%%D" (
        echo Creating %%D directory... | %TEE%
        mkdir "%%D"
        if !ERRORLEVEL! NEQ 0 (
            echo Failed to create %%D directory. | %TEE%
            echo Press any key to exit... | %TEE%
            pause > nul
            exit /b 1
        )
    )
)

REM Create a basic config file if it doesn't exist
if not exist "config\config.json" (
    echo Creating default configuration... | %TEE%
    (
        echo {
        echo   "github_copilot_url": "https://copilot.github.com/",
        echo   "browser_user_data_dir": "browser_data",
        echo   "screenshot_dir": "screenshots",
        echo   "log_file": "logs/agent.log",
        echo   "headless": false,
        echo   "debug_mode": true,
        echo   "github_login_required": true
        echo }
    ) > config\config.json
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to create configuration file. | %TEE%
        echo Press any key to exit... | %TEE%
        pause > nul
        exit /b 1
    )
    echo Configuration file created at config\config.json | %TEE%
    echo NOTE: You will need to log in to GitHub Copilot when the browser opens. | %TEE%
)

REM Install Python dependencies with specific versions
echo Installing Python dependencies... | %TEE%
echo Running: python -m pip install --upgrade pip | %TEE%
python -m pip install --upgrade pip
set EXIT_CODE=%ERRORLEVEL%
echo pip upgrade completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to upgrade pip. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

echo Running: python -m pip install selenium==4.15.2 webdriver-manager==4.0.1 pyautogui==0.9.54 requests==2.31.0 flask==3.0.0 | %TEE%
python -m pip install selenium==4.15.2 webdriver-manager==4.0.1 pyautogui==0.9.54 requests==2.31.0 flask==3.0.0
set EXIT_CODE=%ERRORLEVEL%
echo Python dependencies installation completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install Python dependencies. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Install Go dependencies
echo Installing Go dependencies... | %TEE%

REM Check if go.mod exists
if not exist "go.mod" (
    echo Initializing Go module... | %TEE%
    go mod init web-integration-agent
    set EXIT_CODE=%ERRORLEVEL%
    echo go mod init completed with exit code %EXIT_CODE% | %TEE%
    
    if %EXIT_CODE% NEQ 0 (
        echo Failed to initialize Go module. | %TEE%
        echo Press any key to exit... | %TEE%
        pause > nul
        exit /b 1
    )
)

REM Install Go dependencies with error checking
echo Installing chromedp dependency... | %TEE%
go get github.com/chromedp/chromedp
set EXIT_CODE=%ERRORLEVEL%
echo chromedp installation completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install chromedp dependency. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

echo Installing gorilla/mux dependency... | %TEE%
go get github.com/gorilla/mux
set EXIT_CODE=%ERRORLEVEL%
echo gorilla/mux installation completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install gorilla/mux dependency. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

echo Installing rs/cors dependency... | %TEE%
go get github.com/rs/cors
set EXIT_CODE=%ERRORLEVEL%
echo rs/cors installation completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install rs/cors dependency. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
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

REM Install specific cdproto packages
echo Installing specific cdproto packages... | %TEE%
echo Running: go get github.com/chromedp/cdproto/runtime/enable | %TEE%
go get github.com/chromedp/cdproto/runtime/enable
set EXIT_CODE=%ERRORLEVEL%
echo cdproto/runtime/enable installation completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install cdproto/runtime/enable dependency. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

echo Running: go get github.com/tidwall/gjson@v1.18.0 | %TEE%
go get github.com/tidwall/gjson@v1.18.0
set EXIT_CODE=%ERRORLEVEL%
echo gjson installation completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to install gjson dependency. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

echo Running go mod tidy... | %TEE%
go mod tidy
set EXIT_CODE=%ERRORLEVEL%
echo go mod tidy completed with exit code %EXIT_CODE% | %TEE%

if %EXIT_CODE% NEQ 0 (
    echo Failed to tidy go modules. | %TEE%
    echo Press any key to exit... | %TEE%
    pause > nul
    exit /b 1
)

REM Build and run the web integration agent for GitHub Copilot
echo Building and starting the Web Copilot OpenAI API compatible endpoint... | %TEE%
echo Running: go run web-integration-agent.go --mode copilot --port 8080 | %TEE%
go run web-integration-agent.go --mode copilot --port 8080
set EXIT_CODE=%ERRORLEVEL%
echo Web Copilot OpenAI API endpoint exited with code %EXIT_CODE% | %TEE%

REM If the program exits, pause to see any error messages
echo. | %TEE%
echo The Web Copilot OpenAI API endpoint has stopped. | %TEE%
echo Press any key to exit... | %TEE%
pause > nul

endlocal

