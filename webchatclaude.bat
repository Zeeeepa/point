@echo off
setlocal enabledelayedexpansion

REM ===================================================
REM WebChatClaude - OpenAI API compatible endpoint
REM Version: 2.0.0
REM ===================================================

REM Initialize configuration variables with defaults
set "PORT=8081"
set "FALLBACK_PORT=8082"
set "MAX_RETRIES=3"
set "HEADLESS=false"
set "LOG_DIR=%~dp0logs"
set "TIMESTAMP=%date:~-4,4%-%date:~-7,2%-%date:~-10,2%_%time:~0,2%-%time:~3,2%-%time:~6,2%"
set "TIMESTAMP=%TIMESTAMP: =0%"
set "SETUP_LOG=%LOG_DIR%\setup_%TIMESTAMP%.log"
set "ERROR_LOG=%LOG_DIR%\error_%TIMESTAMP%.log"
set "SERVICE_LOG=%LOG_DIR%\service_%TIMESTAMP%.log"
set "PIP_LOG=%LOG_DIR%\pip_%TIMESTAMP%.log"
set "GO_LOG=%LOG_DIR%\go_%TIMESTAMP%.log"
set "PLAYWRIGHT_LOG=%LOG_DIR%\playwright_%TIMESTAMP%.log"

REM Process command-line arguments
:parse_args
if "%~1"=="" goto :end_parse_args
if /i "%~1"=="--port" (
    set "PORT=%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--fallback-port" (
    set "FALLBACK_PORT=%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--headless" (
    set "HEADLESS=true"
    shift
    goto :parse_args
)
if /i "%~1"=="--max-retries" (
    set "MAX_RETRIES=%~2"
    shift
    shift
    goto :parse_args
)
if /i "%~1"=="--help" (
    echo WebChatClaude - OpenAI API compatible endpoint for Claude
    echo.
    echo Usage: webchatclaude.bat [options]
    echo.
    echo Options:
    echo   --port PORT            Set the primary port (default: 8081)
    echo   --fallback-port PORT   Set the fallback port (default: 8082)
    echo   --headless             Run in headless mode
    echo   --max-retries N        Set maximum retry attempts (default: 3)
    echo   --help                 Display this help message
    echo.
    exit /b 0
)
shift
goto :parse_args
:end_parse_args

REM Create logs directory if it doesn't exist
if not exist "%LOG_DIR%" (
    mkdir "%LOG_DIR%" 2>nul
    echo [%date% %time%] Created logs directory: %LOG_DIR% >> "%SETUP_LOG%"
)

REM Log startup information
echo [%date% %time%] ===================================================>> "%SETUP_LOG%"
echo [%date% %time%] Starting Web Claude OpenAI API compatible endpoint...>> "%SETUP_LOG%"
echo [%date% %time%] Configuration:>> "%SETUP_LOG%"
echo [%date% %time%]   - Primary Port: %PORT%>> "%SETUP_LOG%"
echo [%date% %time%]   - Fallback Port: %FALLBACK_PORT%>> "%SETUP_LOG%"
echo [%date% %time%]   - Headless Mode: %HEADLESS%>> "%SETUP_LOG%"
echo [%date% %time%]   - Max Retries: %MAX_RETRIES%>> "%SETUP_LOG%"
echo [%date% %time%] ===================================================>> "%SETUP_LOG%"

echo ===================================================
echo Starting Web Claude OpenAI API compatible endpoint...
echo Configuration:
echo   - Primary Port: %PORT%
echo   - Fallback Port: %FALLBACK_PORT%
echo   - Headless Mode: %HEADLESS%
echo   - Max Retries: %MAX_RETRIES%
echo ===================================================
echo.

REM Call common installation script with Python
echo [%date% %time%] Running common installation script with Python...>> "%SETUP_LOG%"
call "%~dp0install-common.bat" with-python > "%PIP_LOG%" 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [%date% %time%] ERROR: Failed to install common dependencies. >> "%ERROR_LOG%"
    echo [%date% %time%] See log for details: %PIP_LOG% >> "%ERROR_LOG%"
    echo ERROR: Failed to install common dependencies.
    echo See log for details: %PIP_LOG%
    pause
    exit /b 1
)
echo [%date% %time%] Common installation completed successfully.>> "%SETUP_LOG%"

REM Check if the ai-web-integration-agent directory exists
if not exist "%~dp0ai-web-integration-agent" (
    echo [%date% %time%] ERROR: ai-web-integration-agent directory not found. >> "%ERROR_LOG%"
    echo ERROR: ai-web-integration-agent directory not found.
    echo Please download the repository from https://github.com/Zeeeepa/point/tree/main/ai-web-integration-agent
    echo or clone it using: git clone https://github.com/Zeeeepa/point.git
    pause
    exit /b 1
)
echo [%date% %time%] ai-web-integration-agent directory found.>> "%SETUP_LOG%"

REM Navigate to the ai-web-integration-agent directory
cd "%~dp0ai-web-integration-agent"
echo [%date% %time%] Changed directory to: %cd%>> "%SETUP_LOG%"

REM Create required directories if they don't exist
for %%D in (config logs screenshots browser_data) do (
    if not exist "%%D" (
        mkdir "%%D" 2>nul
        echo [%date% %time%] Created directory: %%D>> "%SETUP_LOG%"
    ) else (
        echo [%date% %time%] Directory already exists: %%D>> "%SETUP_LOG%"
    )
)

REM Create a basic config file if it doesn't exist
if not exist "config\config.json" (
    echo [%date% %time%] Creating default configuration file...>> "%SETUP_LOG%"
    (
        echo {
        echo   "claude_url": "https://claude.ai/",
        echo   "browser_user_data_dir": "browser_data",
        echo   "screenshot_dir": "screenshots",
        echo   "log_file": "logs/agent.log",
        echo   "headless": %HEADLESS%,
        echo   "debug_mode": true,
        echo   "claude_login_required": true
        echo }
    ) > config\config.json
    echo [%date% %time%] Configuration file created at config\config.json>> "%SETUP_LOG%"
    echo Configuration file created at config\config.json
    echo NOTE: You will need to log in to Claude when the browser opens.
) else (
    echo [%date% %time%] Configuration file already exists.>> "%SETUP_LOG%"
    
    REM Update headless setting in config if specified
    if "%HEADLESS%"=="true" (
        echo [%date% %time%] Updating headless mode in configuration...>> "%SETUP_LOG%"
        powershell -Command "(Get-Content config\config.json) -replace '\"headless\":\s*(false|true)', '\"headless\": true' | Set-Content config\config.json"
    )
)

REM Install Python dependencies with specific versions
echo [%date% %time%] Installing Python dependencies...>> "%SETUP_LOG%"
echo Installing Python dependencies...
python -m pip install --upgrade pip >> "%PIP_LOG%" 2>&1
python -m pip install selenium==4.15.2 webdriver-manager==4.0.1 pyautogui==0.9.54 requests==2.31.0 flask==3.0.0 >> "%PIP_LOG%" 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [%date% %time%] ERROR: Failed to install Python dependencies. >> "%ERROR_LOG%"
    echo [%date% %time%] See log for details: %PIP_LOG% >> "%ERROR_LOG%"
    echo ERROR: Failed to install Python dependencies.
    echo See log for details: %PIP_LOG%
    pause
    exit /b 1
)
echo [%date% %time%] Python dependencies installed successfully.>> "%SETUP_LOG%"

REM Install Go dependencies
echo [%date% %time%] Installing Go dependencies...>> "%SETUP_LOG%"
echo Installing Go dependencies...

REM Check if go.mod exists
if not exist "go.mod" (
    echo [%date% %time%] Initializing Go module...>> "%SETUP_LOG%"
    echo Initializing Go module...
    go mod init web-integration-agent >> "%GO_LOG%" 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo [%date% %time%] ERROR: Failed to initialize Go module. >> "%ERROR_LOG%"
        echo [%date% %time%] See log for details: %GO_LOG% >> "%ERROR_LOG%"
        echo ERROR: Failed to initialize Go module.
        echo See log for details: %GO_LOG%
        pause
        exit /b 1
    )
    echo [%date% %time%] Go module initialized successfully.>> "%SETUP_LOG%"
)

REM Install Go dependencies with error checking and retry logic
set "GO_DEPS=github.com/chromedp/chromedp github.com/gorilla/mux github.com/rs/cors github.com/chromedp/cdproto/runtime/enable github.com/tidwall/gjson@v1.18.0"

for %%D in (%GO_DEPS%) do (
    set "RETRY_COUNT=0"
    :retry_go_dep
    echo [%date% %time%] Installing %%D...>> "%SETUP_LOG%"
    echo Installing %%D...
    go get %%D >> "%GO_LOG%" 2>&1
    if %ERRORLEVEL% NEQ 0 (
        set /a "RETRY_COUNT+=1"
        if !RETRY_COUNT! LSS %MAX_RETRIES% (
            echo [%date% %time%] Retry !RETRY_COUNT!/%MAX_RETRIES%: Failed to install %%D, retrying...>> "%ERROR_LOG%"
            echo Retry !RETRY_COUNT!/%MAX_RETRIES%: Failed to install %%D, retrying...
            timeout /t 2 > nul
            goto :retry_go_dep
        ) else (
            echo [%date% %time%] ERROR: Failed to install %%D after %MAX_RETRIES% attempts. >> "%ERROR_LOG%"
            echo [%date% %time%] See log for details: %GO_LOG% >> "%ERROR_LOG%"
            echo ERROR: Failed to install %%D after %MAX_RETRIES% attempts.
            echo See log for details: %GO_LOG%
            pause
            exit /b 1
        )
    )
    echo [%date% %time%] Successfully installed %%D.>> "%SETUP_LOG%"
)

REM Call Playwright installation script
echo [%date% %time%] Installing Playwright dependencies...>> "%SETUP_LOG%"
echo Installing Playwright dependencies...
call "%~dp0install-playwright.bat" > "%PLAYWRIGHT_LOG%" 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [%date% %time%] ERROR: Failed to install Playwright dependencies. >> "%ERROR_LOG%"
    echo [%date% %time%] See log for details: %PLAYWRIGHT_LOG% >> "%ERROR_LOG%"
    echo ERROR: Failed to install Playwright dependencies.
    echo See log for details: %PLAYWRIGHT_LOG%
    pause
    exit /b 1
)
echo [%date% %time%] Playwright dependencies installed successfully.>> "%SETUP_LOG%"

echo [%date% %time%] Running go mod tidy...>> "%SETUP_LOG%"
echo Running go mod tidy...
go mod tidy >> "%GO_LOG%" 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [%date% %time%] ERROR: Failed to tidy go modules. >> "%ERROR_LOG%"
    echo [%date% %time%] See log for details: %GO_LOG% >> "%ERROR_LOG%"
    echo ERROR: Failed to tidy go modules.
    echo See log for details: %GO_LOG%
    pause
    exit /b 1
)
echo [%date% %time%] Go modules tidied successfully.>> "%SETUP_LOG%"

REM Check if primary port is available
set "PORT_TO_USE=%PORT%"
echo [%date% %time%] Checking if port %PORT% is available...>> "%SETUP_LOG%"
netstat -an | find ":%PORT% " > nul
if %ERRORLEVEL% EQU 0 (
    echo [%date% %time%] WARNING: Port %PORT% is already in use. Trying fallback port %FALLBACK_PORT%...>> "%ERROR_LOG%"
    echo WARNING: Port %PORT% is already in use. Trying fallback port %FALLBACK_PORT%...
    set "PORT_TO_USE=%FALLBACK_PORT%"
    
    REM Check if fallback port is available
    netstat -an | find ":%FALLBACK_PORT% " > nul
    if %ERRORLEVEL% EQU 0 (
        echo [%date% %time%] ERROR: Both primary port %PORT% and fallback port %FALLBACK_PORT% are in use. >> "%ERROR_LOG%"
        echo ERROR: Both primary port %PORT% and fallback port %FALLBACK_PORT% are in use.
        echo Please specify different ports using --port and --fallback-port options.
        pause
        exit /b 1
    )
)

REM Build and run the web integration agent for Claude with retry logic
set "RETRY_COUNT=0"
:retry_service
echo [%date% %time%] Starting Web Claude OpenAI API compatible endpoint on port %PORT_TO_USE%...>> "%SERVICE_LOG%"
echo Building and starting the Web Claude OpenAI API compatible endpoint on port %PORT_TO_USE%...

REM Construct the command with appropriate flags
set "CMD=go run web-integration-agent.go --mode claude --port %PORT_TO_USE%"
if "%HEADLESS%"=="true" (
    set "CMD=%CMD% --headless"
)

%CMD% >> "%SERVICE_LOG%" 2>&1
set "EXIT_CODE=%ERRORLEVEL%"

if %EXIT_CODE% NEQ 0 (
    set /a "RETRY_COUNT+=1"
    echo [%date% %time%] ERROR: Service failed with exit code %EXIT_CODE%. >> "%ERROR_LOG%"
    
    if !RETRY_COUNT! LSS %MAX_RETRIES% (
        echo [%date% %time%] Retry !RETRY_COUNT!/%MAX_RETRIES%: Service failed, retrying... >> "%ERROR_LOG%"
        echo Retry !RETRY_COUNT!/%MAX_RETRIES%: Service failed, retrying...
        
        REM If we've tried normal mode and it failed, try headless mode
        if "%HEADLESS%"=="false" (
            if !RETRY_COUNT! GEQ 2 (
                echo [%date% %time%] Switching to headless mode as fallback... >> "%SERVICE_LOG%"
                echo Switching to headless mode as fallback...
                set "HEADLESS=true"
                
                REM Update config file for headless mode
                powershell -Command "(Get-Content config\config.json) -replace '\"headless\":\s*(false|true)', '\"headless\": true' | Set-Content config\config.json"
            )
        )
        
        timeout /t 5 > nul
        goto :retry_service
    ) else (
        echo [%date% %time%] ERROR: Service failed after %MAX_RETRIES% attempts. >> "%ERROR_LOG%"
        echo ERROR: Service failed after %MAX_RETRIES% attempts.
        echo See logs for details:
        echo - Setup Log: %SETUP_LOG%
        echo - Error Log: %ERROR_LOG%
        echo - Service Log: %SERVICE_LOG%
        pause
        exit /b 1
    )
)

REM If the program exits normally
echo [%date% %time%] The Web Claude OpenAI API endpoint has stopped normally.>> "%SERVICE_LOG%"
echo.
echo The Web Claude OpenAI API endpoint has stopped.
echo Logs are available at:
echo - Setup Log: %SETUP_LOG%
echo - Service Log: %SERVICE_LOG%
pause

