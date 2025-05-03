@echo off
setlocal enabledelayedexpansion

REM ===================================================
REM Cursor OpenAI API compatible endpoint launcher
REM Version: 1.1.0
REM ===================================================

REM Set up logging
set "LOG_DIR=%~dp0logs"
set "LOG_FILE=%LOG_DIR%\cursoropenapi_%date:~-4,4%%date:~-7,2%%date:~-10,2%_%time:~0,2%%time:~3,2%%time:~6,2%.log"
set "LOG_FILE=%LOG_FILE: =0%"

REM Create logs directory if it doesn't exist
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

REM Function to log messages
call :log "====================================================="
call :log "Starting Cursor OpenAI API compatible endpoint..."
call :log "====================================================="
call :log "Timestamp: %date% %time%"
call :log "Working directory: %~dp0"

echo =====================================================
echo Starting Cursor OpenAI API compatible endpoint...
echo =====================================================
echo.
echo [INFO] Logs will be saved to: %LOG_FILE%
echo.

REM Check if we're running with admin privileges
net session >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] Not running with administrator privileges.
    echo [WARNING] Some installation steps might fail.
    call :log "WARNING: Not running with administrator privileges"
)

REM Call common installation script with retry mechanism
call :log "Calling install-common.bat..."
call "%~dp0install-common.bat"
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install common dependencies. Retrying...
    call :log "ERROR: Failed to install common dependencies. Retrying..."
    timeout /t 3 >nul
    call "%~dp0install-common.bat"
    if %ERRORLEVEL% NEQ 0 (
        echo [CRITICAL] Failed to install common dependencies after retry.
        call :log "CRITICAL: Failed to install common dependencies after retry"
        echo Please check the log file for details: %LOG_FILE%
        echo You can try running this script as administrator or manually install Go from https://go.dev/dl/
        pause
        exit /b 1
    )
)
call :log "Common dependencies installed successfully"

REM Check if the chatgpt-adapter-main directory exists
if not exist "%~dp0chatgpt-adapter-main" (
    echo [ERROR] chatgpt-adapter-main directory not found.
    call :log "ERROR: chatgpt-adapter-main directory not found"
    
    REM Fallback: Try to clone the repository
    echo [INFO] Attempting to download the repository...
    call :log "Attempting to download the repository..."
    
    where git >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [INFO] Using git to clone the repository...
        call :log "Using git to clone the repository"
        git clone https://github.com/Zeeeepa/point.git temp_point
        if %ERRORLEVEL% NEQ 0 (
            echo [ERROR] Git clone failed.
            call :log "ERROR: Git clone failed"
            goto :download_fallback
        )
        
        if exist "temp_point\chatgpt-adapter-main" (
            echo [INFO] Copying chatgpt-adapter-main from cloned repository...
            call :log "Copying chatgpt-adapter-main from cloned repository"
            xcopy /E /I /Y "temp_point\chatgpt-adapter-main" "%~dp0chatgpt-adapter-main"
            rmdir /S /Q temp_point
        ) else (
            echo [ERROR] chatgpt-adapter-main not found in cloned repository.
            call :log "ERROR: chatgpt-adapter-main not found in cloned repository"
            rmdir /S /Q temp_point
            goto :download_fallback
        )
    ) else (
        :download_fallback
        echo [INFO] Attempting to download using PowerShell...
        call :log "Attempting to download using PowerShell"
        powershell -Command "& {try { Invoke-WebRequest -Uri 'https://github.com/Zeeeepa/point/archive/refs/heads/main.zip' -OutFile 'point.zip'; Expand-Archive -Path 'point.zip' -DestinationPath '.'; if (Test-Path '.\point-main\chatgpt-adapter-main') { Copy-Item -Path '.\point-main\chatgpt-adapter-main' -Destination '.' -Recurse; }; Remove-Item -Path 'point.zip'; Remove-Item -Path '.\point-main' -Recurse; } catch { Write-Host 'PowerShell download failed'; exit 1 }}"
        if %ERRORLEVEL% NEQ 0 (
            echo [CRITICAL] All download attempts failed.
            call :log "CRITICAL: All download attempts failed"
            echo Please manually download the repository from https://github.com/Zeeeepa/point/tree/main/chatgpt-adapter-main
            echo or clone it using: git clone https://github.com/Zeeeepa/point.git
            pause
            exit /b 1
        )
    )
    
    REM Check if download was successful
    if not exist "%~dp0chatgpt-adapter-main" (
        echo [CRITICAL] Failed to obtain chatgpt-adapter-main directory.
        call :log "CRITICAL: Failed to obtain chatgpt-adapter-main directory"
        echo Please manually download the repository from https://github.com/Zeeeepa/point/tree/main/chatgpt-adapter-main
        echo or clone it using: git clone https://github.com/Zeeeepa/point.git
        pause
        exit /b 1
    )
    
    echo [SUCCESS] Successfully downloaded chatgpt-adapter-main.
    call :log "SUCCESS: Successfully downloaded chatgpt-adapter-main"
)

REM Navigate to the chatgpt-adapter-main directory
cd "%~dp0chatgpt-adapter-main"
call :log "Changed directory to chatgpt-adapter-main"

REM Check if dependencies are installed
if not exist "go.sum" (
    echo [INFO] Installing dependencies...
    call :log "Installing dependencies..."
    go mod download
    if %ERRORLEVEL% NEQ 0 (
        echo [ERROR] Failed to download dependencies. Retrying with proxy...
        call :log "ERROR: Failed to download dependencies. Retrying with proxy..."
        
        REM Try with GOPROXY
        set GOPROXY=https://goproxy.io,direct
        echo [INFO] Set GOPROXY to https://goproxy.io,direct
        call :log "Set GOPROXY to https://goproxy.io,direct"
        
        go mod download
        if %ERRORLEVEL% NEQ 0 (
            echo [CRITICAL] Failed to download dependencies even with proxy.
            call :log "CRITICAL: Failed to download dependencies even with proxy"
            echo Please check your internet connection and try again.
            pause
            exit /b 1
        )
    )
    call :log "Dependencies downloaded successfully"
)

REM Call Playwright installation script with retry
call :log "Calling install-playwright.bat..."
call "%~dp0install-playwright.bat"
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install Playwright dependencies. Retrying...
    call :log "ERROR: Failed to install Playwright dependencies. Retrying..."
    timeout /t 3 >nul
    call "%~dp0install-playwright.bat"
    if %ERRORLEVEL% NEQ 0 (
        echo [WARNING] Failed to install Playwright dependencies after retry.
        call :log "WARNING: Failed to install Playwright dependencies after retry"
        echo [INFO] Continuing without Playwright. Some features may not work.
        call :log "Continuing without Playwright. Some features may not work"
    ) else {
        call :log "Playwright dependencies installed successfully on second attempt"
    }
) else {
    call :log "Playwright dependencies installed successfully"
}

REM Fix missing go.sum entries
echo [INFO] Fixing go.sum entries...
call :log "Fixing go.sum entries..."
go get github.com/tidwall/gjson@v1.18.0
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] Failed to get gjson dependency.
    call :log "WARNING: Failed to get gjson dependency"
    echo [INFO] Trying alternative approach...
    call :log "Trying alternative approach..."
    
    REM Try with direct go.mod edit
    echo replace github.com/tidwall/gjson => github.com/tidwall/gjson v1.18.0 >> go.mod
    call :log "Added replace directive to go.mod"
)

go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] Failed to fix go.sum entries.
    call :log "WARNING: Failed to fix go.sum entries"
    echo [INFO] Continuing anyway, but you may encounter issues.
    call :log "Continuing anyway, but you may encounter issues"
) else {
    call :log "go.mod tidy completed successfully"
}

REM Create logs directory if it doesn't exist
if not exist "logs" mkdir logs
call :log "Ensured logs directory exists in chatgpt-adapter-main"

REM Set default port and check if it's available
set PORT=8080
call :log "Default port set to 8080"

REM Check if port is already in use
netstat -ano | findstr ":%PORT%" >nul
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Port %PORT% is already in use.
    call :log "WARNING: Port %PORT% is already in use"
    
    REM Try alternative ports
    for %%p in (8081 8082 8083 8084 8085) do (
        netstat -ano | findstr ":%%p" >nul
        if %ERRORLEVEL% NEQ 0 (
            set PORT=%%p
            echo [INFO] Using alternative port: %%p
            call :log "Using alternative port: %%p"
            goto :port_found
        )
    )
    
    echo [WARNING] All alternative ports are in use. Using default port anyway.
    call :log "WARNING: All alternative ports are in use. Using default port anyway"
    set PORT=8080
)
:port_found

REM Build and run the adapter with proper error handling
echo [INFO] Building and starting the Cursor OpenAI API compatible endpoint...
call :log "Building and starting the Cursor OpenAI API compatible endpoint on port %PORT%..."

REM Run with proper error handling
echo [INFO] Starting server on port %PORT%...
call :log "Starting server on port %PORT%..."
echo [INFO] API will be available at: http://localhost:%PORT%/v1/chat/completions
call :log "API will be available at: http://localhost:%PORT%/v1/chat/completions"
echo.
echo [INFO] Press Ctrl+C to stop the server
echo.

REM Run the adapter with the specified port
go run main.go --port %PORT% --mode cursor 2>> "%LOG_FILE%"

REM If the program exits, check the error code
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] The Cursor OpenAI API endpoint has stopped with an error.
    call :log "ERROR: The Cursor OpenAI API endpoint has stopped with error code %ERRORLEVEL%"
    echo [INFO] Check the log file for details: %LOG_FILE%
) else (
    echo.
    echo [INFO] The Cursor OpenAI API endpoint has stopped normally.
    call :log "INFO: The Cursor OpenAI API endpoint has stopped normally"
)

pause
exit /b 0

REM Log function
:log
echo %~1 >> "%LOG_FILE%"
goto :eof

