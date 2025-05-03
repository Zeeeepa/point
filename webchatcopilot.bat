@echo off
setlocal enabledelayedexpansion

REM ===================================================
REM Improved Web Copilot OpenAI API compatible endpoint
REM With enhanced logging and fallback mechanisms
REM ===================================================

REM Create logs directory if it doesn't exist
if not exist "%~dp0logs" mkdir "%~dp0logs"

REM Initialize log file with timestamp
set LOG_FILE="%~dp0logs\webchatcopilot_%date:~-4,4%%date:~-7,2%%date:~-10,2%_%time:~0,2%%time:~3,2%%time:~6,2%.log"
set LOG_FILE=%LOG_FILE: =0%
echo [%date% %time%] Starting Web Copilot OpenAI API compatible endpoint setup > %LOG_FILE%

REM Function to log messages to both console and log file
:log
echo [%date% %time%] %~1
echo [%date% %time%] %~1 >> %LOG_FILE%
goto :eof

REM Display welcome message
echo ===================================================
echo Starting Web Copilot OpenAI API compatible endpoint...
echo ===================================================
echo.
call :log "Starting Web Copilot OpenAI API compatible endpoint..."

REM Check if the service is already running
call :log "Checking if service is already running..."
tasklist /FI "IMAGENAME eq web-integration-agent.exe" 2>NUL | find /I /N "web-integration-agent.exe">NUL
if "%ERRORLEVEL%"=="0" (
    call :log "Web integration agent is already running"
    choice /C YN /M "Service is already running. Do you want to restart it"
    if errorlevel 2 (
        call :log "User chose not to restart. Exiting."
        exit /b 0
    ) else (
        call :log "Restarting service..."
        taskkill /F /IM web-integration-agent.exe >NUL 2>&1
        if errorlevel 1 (
            call :log "WARNING: Failed to kill existing process. It may be running with different permissions."
        )
    )
)

REM Call common installation script with Python
call :log "Calling install-common.bat with Python..."
call "%~dp0install-common.bat" with-python
if %ERRORLEVEL% NEQ 0 (
    call :log "ERROR: Failed to install common dependencies."
    call :log "Attempting fallback installation..."
    call :install_common_fallback
    if %ERRORLEVEL% NEQ 0 (
        call :log "CRITICAL: All installation methods failed. Please check your system configuration."
        echo Failed to install common dependencies. See log file for details: %LOG_FILE%
        pause
        exit /b 1
    )
)

REM Check if the ai-web-integration-agent directory exists
call :log "Checking for ai-web-integration-agent directory..."
if not exist "%~dp0ai-web-integration-agent" (
    call :log "ERROR: ai-web-integration-agent directory not found."
    call :log "Attempting to clone repository..."
    
    REM Try to clone the repository as a fallback
    where git >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        call :log "Git found, attempting to clone repository..."
        git clone https://github.com/Zeeeepa/point.git "%~dp0point_temp"
        if %ERRORLEVEL% EQU 0 (
            if exist "%~dp0point_temp\ai-web-integration-agent" (
                call :log "Successfully cloned repository, copying ai-web-integration-agent directory..."
                xcopy /E /I "%~dp0point_temp\ai-web-integration-agent" "%~dp0ai-web-integration-agent"
                rmdir /S /Q "%~dp0point_temp"
            ) else (
                call :log "ERROR: ai-web-integration-agent directory not found in cloned repository."
                rmdir /S /Q "%~dp0point_temp"
            )
        ) else (
            call :log "ERROR: Failed to clone repository."
        )
    ) else (
        call :log "Git not found, cannot clone repository."
    )
    
    REM Check if the directory exists after fallback attempt
    if not exist "%~dp0ai-web-integration-agent" (
        call :log "CRITICAL: Could not find or create ai-web-integration-agent directory."
        echo Error: ai-web-integration-agent directory not found.
        echo Please download the repository from https://github.com/Zeeeepa/point/tree/main/ai-web-integration-agent
        echo or clone it using: git clone https://github.com/Zeeeepa/point.git
        echo See log file for details: %LOG_FILE%
        pause
        exit /b 1
    )
)

REM Navigate to the ai-web-integration-agent directory
call :log "Navigating to ai-web-integration-agent directory..."
cd "%~dp0ai-web-integration-agent"

REM Create required directories if they don't exist
call :log "Creating required directories..."
for %%D in (config logs screenshots browser_data) do (
    if not exist "%%D" (
        call :log "Creating %%D directory..."
        mkdir "%%D"
    )
)

REM Create a basic config file if it doesn't exist
if not exist "config\config.json" (
    call :log "Creating default configuration file..."
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
    call :log "Configuration file created at config\config.json"
    call :log "NOTE: You will need to log in to GitHub Copilot when the browser opens."
) else (
    call :log "Configuration file already exists."
)

REM Backup existing config file
if exist "config\config.json" (
    call :log "Backing up existing configuration..."
    copy "config\config.json" "config\config.json.bak" >NUL
    if errorlevel 1 (
        call :log "WARNING: Failed to backup configuration file."
    )
)

REM Install Python dependencies with specific versions
call :log "Installing Python dependencies..."
python -m pip install --upgrade pip
if errorlevel 1 (
    call :log "WARNING: Failed to upgrade pip. Continuing with existing version."
)

REM Try to install all dependencies at once first
call :log "Installing Python packages..."
python -m pip install selenium==4.15.2 webdriver-manager==4.0.1 pyautogui==0.9.54 requests==2.31.0 flask==3.0.0
if errorlevel 1 (
    call :log "ERROR: Failed to install Python dependencies as a group."
    call :log "Attempting to install dependencies individually..."
    
    REM Try installing dependencies one by one as fallback
    set FAILED_DEPS=0
    for %%P in (selenium==4.15.2 webdriver-manager==4.0.1 pyautogui==0.9.54 requests==2.31.0 flask==3.0.0) do (
        call :log "Installing %%P..."
        python -m pip install %%P
        if errorlevel 1 (
            call :log "WARNING: Failed to install %%P"
            set /a FAILED_DEPS+=1
        )
    )
    
    if !FAILED_DEPS! GTR 0 (
        call :log "WARNING: !FAILED_DEPS! Python dependencies failed to install."
        call :log "Attempting to continue with available packages..."
    )
)

REM Install Go dependencies
call :log "Installing Go dependencies..."

REM Check if go.mod exists
if not exist "go.mod" (
    call :log "Initializing Go module..."
    go mod init web-integration-agent
    if errorlevel 1 (
        call :log "ERROR: Failed to initialize Go module."
        call :log "Checking Go installation..."
        where go >nul 2>&1
        if errorlevel 1 (
            call :log "ERROR: Go is not installed or not in PATH."
            echo Failed to initialize Go module. Go may not be installed correctly.
            echo See log file for details: %LOG_FILE%
            pause
            exit /b 1
        ) else (
            call :log "Go is installed but module initialization failed."
            call :log "Attempting to continue..."
        )
    )
)

REM Install Go dependencies with error checking and retry
call :install_go_dependency "github.com/chromedp/chromedp" "chromedp"
call :install_go_dependency "github.com/gorilla/mux" "gorilla/mux"
call :install_go_dependency "github.com/rs/cors" "rs/cors"

REM Call Playwright installation script
call :log "Calling install-playwright.bat..."
call "%~dp0install-playwright.bat"
if errorlevel 1 (
    call :log "ERROR: Failed to install Playwright dependencies."
    call :log "Attempting fallback installation..."
    call :install_playwright_fallback
    if errorlevel 1 (
        call :log "CRITICAL: All Playwright installation methods failed."
        echo Failed to install Playwright dependencies. See log file for details: %LOG_FILE%
        pause
        exit /b 1
    )
)

REM Install specific cdproto packages
call :log "Installing specific cdproto packages..."
call :install_go_dependency "github.com/chromedp/cdproto/runtime/enable" "cdproto/runtime/enable"
call :install_go_dependency "github.com/tidwall/gjson@v1.18.0" "gjson"

call :log "Running go mod tidy..."
go mod tidy
if errorlevel 1 (
    call :log "WARNING: Failed to tidy go modules. Attempting to continue..."
)

REM Build and run the web integration agent for GitHub Copilot
call :log "Building and starting the Web Copilot OpenAI API compatible endpoint..."
call :log "Command: go run web-integration-agent.go --mode copilot --port 8080"

REM Start the service with retry logic
set MAX_RETRIES=3
set RETRY_COUNT=0

:start_service
set /a RETRY_COUNT+=1
call :log "Starting service (attempt !RETRY_COUNT! of %MAX_RETRIES%)..."

go run web-integration-agent.go --mode copilot --port 8080
set EXIT_CODE=%ERRORLEVEL%

if %EXIT_CODE% NEQ 0 (
    call :log "ERROR: Service exited with code %EXIT_CODE%"
    
    if !RETRY_COUNT! LSS %MAX_RETRIES% (
        call :log "Waiting 5 seconds before retry..."
        timeout /t 5 /nobreak >NUL
        goto start_service
    ) else (
        call :log "CRITICAL: Service failed to start after %MAX_RETRIES% attempts."
    )
)

REM If the program exits, log and pause to see any error messages
call :log "The Web Copilot OpenAI API endpoint has stopped with exit code %EXIT_CODE%."
echo.
echo The Web Copilot OpenAI API endpoint has stopped.
echo See log file for details: %LOG_FILE%
pause
exit /b %EXIT_CODE%

REM ===================================================
REM Helper functions
REM ===================================================

:install_go_dependency
REM Parameters: %~1 = dependency path, %~2 = friendly name
call :log "Installing %~2 dependency..."
set MAX_ATTEMPTS=3
set ATTEMPT=0

:retry_go_dep
set /a ATTEMPT+=1
call :log "Attempt !ATTEMPT! of %MAX_ATTEMPTS% for %~2..."
go get %~1
if errorlevel 1 (
    if !ATTEMPT! LSS %MAX_ATTEMPTS% (
        call :log "Failed to install %~2 dependency. Retrying in 5 seconds..."
        timeout /t 5 /nobreak >NUL
        goto retry_go_dep
    ) else (
        call :log "ERROR: Failed to install %~2 dependency after %MAX_ATTEMPTS% attempts."
        call :log "Attempting to continue without this dependency..."
    )
) else (
    call :log "Successfully installed %~2 dependency."
)
goto :eof

:install_common_fallback
call :log "Attempting fallback installation of common dependencies..."
REM Try alternative installation methods for Go and Python
where choco >nul 2>&1
if errorlevel 0 (
    call :log "Chocolatey found, attempting to install dependencies..."
    choco install golang -y
    choco install python -y
    exit /b %ERRORLEVEL%
) else (
    call :log "Chocolatey not found, cannot use fallback installation."
    exit /b 1
)
goto :eof

:install_playwright_fallback
call :log "Attempting fallback installation of Playwright..."
REM Try alternative installation method for Playwright
call :log "Attempting to install Playwright using npm..."
where npm >nul 2>&1
if errorlevel 0 (
    call :log "npm found, attempting to install playwright..."
    npm install playwright
    exit /b %ERRORLEVEL%
) else (
    call :log "npm not found, cannot use fallback installation."
    exit /b 1
)
goto :eof

