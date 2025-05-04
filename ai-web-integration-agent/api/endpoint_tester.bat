@echo off
setlocal enabledelayedexpansion

:: Streamlined script for testing Claude, Copilot, and Cursor endpoints on Windows

:: Color codes for Windows console
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "PURPLE=[95m"
set "CYAN=[96m"
set "NC=[0m"
set "BOLD=[1m"

:: Function to print colored text
:print_color
echo %~1%~2%NC%
exit /b

:: Function to print header
:print_header
echo.
echo ================================================================================
call :print_color "%PURPLE%%BOLD%" "  %~1"
echo ================================================================================
echo.
exit /b

:: Function to print status
:print_status
if "%~2"=="success" (
    call :print_color "%GREEN%" "✓ %~1"
) else (
    call :print_color "%RED%" "✗ %~1"
)
exit /b

:: Function to print info
:print_info
call :print_color "%BLUE%" "ℹ %~1"
exit /b

:: Function to print warning
:print_warning
call :print_color "%YELLOW%" "⚠ %~1"
exit /b

:: Function to check if a port is in use
:is_port_in_use
netstat -ano | findstr /C:":%~1 " >nul
if %ERRORLEVEL% equ 0 (
    exit /b 0
) else (
    exit /b 1
)

:: Function to start the API server
:start_server
set "port=%~1"
set "service=%~2"

:: Check if port is already in use
call :is_port_in_use %port%
if %ERRORLEVEL% equ 0 (
    call :print_warning "Port %port% is already in use. This might be another instance of the API server."
    set /p response="Do you want to continue anyway? (y/n): "
    if /i not "!response!"=="y" (
        call :print_info "Exiting..."
        exit /b 1
    )
)

call :print_info "Starting API server for %service% on port %port%..."

:: Determine the correct path to the API server executable
set "script_dir=%~dp0"
set "server_path=%script_dir%web-integration-api.exe"

:: Check if the server executable exists
if not exist "%server_path%" (
    :: Try to build it
    call :print_info "Server executable not found. Attempting to build it..."
    where go >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        pushd "%script_dir%"
        go build -o web-integration-api.exe .
        if %ERRORLEVEL% equ 0 (
            call :print_status "Server built successfully" "success"
        ) else (
            call :print_status "Failed to build server" "failure"
            call :print_info "Please build the server manually using 'go build -o web-integration-api.exe .'"
            popd
            exit /b 1
        )
        popd
    ) else (
        call :print_status "Go compiler not found. Please install Go and build the server manually." "failure"
        exit /b 1
    )
)

:: Create a config file for the server
echo {> "%script_dir%config.json"
echo   "port": %port%,>> "%script_dir%config.json"
echo   "host": "0.0.0.0",>> "%script_dir%config.json"
echo   "claude_url": "https://claude.ai/chat",>> "%script_dir%config.json"
echo   "github_copilot_url": "https://github.com/features/copilot",>> "%script_dir%config.json"
echo   "browser_user_data_dir": "~/.browser-agent",>> "%script_dir%config.json"
echo   "screenshot_dir": "./screenshots",>> "%script_dir%config.json"
echo   "log_file": "./api-server.log",>> "%script_dir%config.json"
echo   "headless": false,>> "%script_dir%config.json"
echo   "debug_mode": true>> "%script_dir%config.json"
echo }>> "%script_dir%config.json"

:: Start the server
start /b "" "%server_path%" --config "%script_dir%config.json" --port "%port%" > nul 2>&1
set "server_pid=%ERRORLEVEL%"

:: Give the server some time to start
timeout /t 3 > nul

:: Check if the server is running
call :is_port_in_use %port%
if %ERRORLEVEL% equ 1 (
    call :print_status "Server failed to start" "failure"
    exit /b 1
)

call :print_status "API server started on port %port%" "success"
echo %server_pid% > "%script_dir%.server_pid"
exit /b 0

:: Function to stop the API server
:stop_server
set "script_dir=%~dp0"
set "pid_file=%script_dir%.server_pid"

if exist "%pid_file%" (
    set /p server_pid=<"%pid_file%"
    call :print_info "Stopping API server..."
    
    :: Try to terminate the server
    taskkill /f /im web-integration-api.exe > nul 2>&1
    
    call :print_status "API server stopped" "success"
    del "%pid_file%" > nul 2>&1
) else (
    call :print_info "No running server found"
)
exit /b

:: Function to test the API endpoint
:test_endpoint
set "port=%~1"
set "service=%~2"
set "model="
set "endpoint="
set "request_data="

if "%service%"=="Claude" (
    set "model=web_claude"
    set "endpoint=/v1/chat/completions"
    set "request_data={\"model\":\"web_claude\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}],\"temperature\":0.7,\"stream\":false}"
) else if "%service%"=="GitHub Copilot" (
    set "model=web_copilot"
    set "endpoint=/v1/completions"
    set "request_data={\"model\":\"web_copilot\",\"prompt\":\"// Say hello\\nfunction greet() {\",\"temperature\":0.7,\"stream\":false}"
) else if "%service%"=="Cursor" (
    set "model=cursor/claude-3.7-sonnet"
    set "endpoint=/v1/chat/completions"
    set "request_data={\"model\":\"cursor/claude-3.7-sonnet\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}],\"temperature\":0.7,\"stream\":false}"
)

call :print_info "Testing %service% endpoint..."

:: Check if curl is available
where curl >nul 2>&1
if %ERRORLEVEL% neq 0 (
    call :print_status "curl command not found. Please install curl to test the endpoint." "failure"
    exit /b 1
)

:: Test the endpoint
set "url=http://localhost:%port%%endpoint%"
curl -s -X POST -H "Content-Type: application/json" -H "Accept: application/json" -d "%request_data%" "%url%" > "%TEMP%\response.json" 2>&1

if %ERRORLEVEL% equ 0 (
    call :print_status "%service% endpoint is working!" "success"
    call :print_info "Response:"
    type "%TEMP%\response.json"
    exit /b 0
) else (
    call :print_status "Endpoint test failed" "failure"
    call :print_info "Response:"
    type "%TEMP%\response.json"
    exit /b 1
)

:: Main menu
:main_menu
cls
call :print_header "AI Web Integration API Tester"

echo Select an AI service to test:
echo 1. Claude
echo 2. GitHub Copilot
echo 3. Cursor
echo 4. Exit

:menu_choice
set /p choice=Enter your choice (1-4): 

if "%choice%"=="1" (
    set "service=Claude"
) else if "%choice%"=="2" (
    set "service=GitHub Copilot"
) else if "%choice%"=="3" (
    set "service=Cursor"
) else if "%choice%"=="4" (
    call :print_info "Exiting..."
    exit /b 0
) else (
    call :print_warning "Invalid choice. Please try again."
    goto menu_choice
)

:: Get port number
:get_port
set /p port_input=Enter port number (default: 8080): 

if "%port_input%"=="" (
    set "port=8080"
) else (
    set "port=%port_input%"
    
    :: Validate port number
    set "valid=true"
    for /f "delims=0123456789" %%i in ("%port%") do set "valid=false"
    
    if "%valid%"=="false" (
        call :print_warning "Port must be a number between 1024 and 65535."
        goto get_port
    )
    
    if %port% lss 1024 (
        call :print_warning "Port must be at least 1024."
        goto get_port
    )
    
    if %port% gtr 65535 (
        call :print_warning "Port must be at most 65535."
        goto get_port
    )
)

set "result=%service%:%port%"
echo %result%
exit /b

:: Main function
:main
setlocal enabledelayedexpansion

:main_loop
    :: Show menu and get choices
    for /f "tokens=1,2 delims=:" %%a in ('call :main_menu') do (
        set "service=%%a"
        set "port=%%b"
    )
    
    :: Start server
    call :start_server "!port!" "!service!"
    if %ERRORLEVEL% neq 0 (
        pause
        goto main_loop
    )
    
    :: Test endpoint
    call :test_endpoint "!port!" "!service!"
    
    :: Ask what to do next
    echo.
    echo What would you like to do next?
    echo 1. Test another service
    echo 2. Exit
    
    :next_choice
    set /p next_choice=Enter your choice (1-2): 
    
    if "!next_choice!"=="1" (
        :: Continue to next iteration
    ) else if "!next_choice!"=="2" (
        call :print_info "Exiting..."
        goto end_main
    ) else (
        call :print_warning "Invalid choice. Please try again."
        goto next_choice
    )
    
    :: Stop the current server
    call :stop_server
    
    goto main_loop

:end_main
:: Stop the server before exiting
call :stop_server
endlocal
exit /b

:: Run the main function
call :main

