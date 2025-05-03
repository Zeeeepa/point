@echo off
REM Common installation script for dependencies
REM This script handles installation of Go and Python dependencies
setlocal enabledelayedexpansion

REM Check if verbose mode is enabled
set VERBOSE=0
if "%~2"=="verbose" set VERBOSE=1

REM Create logs directory if it doesn't exist
if not exist "%~dp0logs" mkdir "%~dp0logs"

REM Get the current date and time for the log filename
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "LOGDATE=%dt:~0,8%-%dt:~8,6%"
set "LOGFILE=%~dp0logs\install-common_%LOGDATE%.log"

REM Start logging
echo Starting common installation script at %TIME% > "%LOGFILE%"
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

if %VERBOSE% EQU 1 (
    echo Verbose mode enabled. Logging to %LOGFILE% | %TEE%
)

REM Check if Go is installed and has minimum version
echo Checking Go installation... | %TEE%
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Go is not installed. Installing Go... | %TEE%
    echo Please wait, this might take a few minutes... | %TEE%
    
    echo Downloading Go installer... | %TEE%
    powershell -Command "& {Invoke-WebRequest -Uri 'https://go.dev/dl/go1.21.0.windows-amd64.msi' -OutFile 'go_installer.msi'}"
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to download Go installer. | %TEE%
        echo Please install Go manually from https://go.dev/dl/ | %TEE%
        exit /b 1
    )
    
    echo Installing Go... | %TEE%
    powershell -Command "& {Start-Process -FilePath 'msiexec.exe' -ArgumentList '/i', 'go_installer.msi', '/quiet' -Wait}"
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to install Go. | %TEE%
        echo Please install Go manually from https://go.dev/dl/ | %TEE%
        exit /b 1
    )
    
    echo Cleaning up installer... | %TEE%
    del "go_installer.msi"
    
    REM Verify Go installation
    where go >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to install Go. Please install it manually from https://go.dev/dl/ | %TEE%
        exit /b 1
    )
    echo Go has been installed successfully. | %TEE%
    
    REM Refresh PATH environment variable
    echo Refreshing PATH environment variable... | %TEE%
    for /f "tokens=2*" %%a in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path') do set "PATH=%%b"
    for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v Path') do set "PATH=!PATH!;%%b"
) else (
    REM Check Go version
    for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
    echo Detected Go version: %GO_VERSION% | %TEE%
    REM Remove 'go' prefix from version string
    set GO_VERSION=%GO_VERSION:go=%
    
    REM Simple version check (requires Go 1.18 or higher)
    for /f "tokens=1,2 delims=." %%a in ("%GO_VERSION%") do (
        set MAJOR=%%a
        set MINOR=%%b
    )
    
    if %MAJOR% LSS 1 (
        echo Go version is too old. Minimum required is 1.18. | %TEE%
        echo Please upgrade Go from https://go.dev/dl/ | %TEE%
        exit /b 1
    ) else (
        if %MAJOR% EQU 1 (
            if %MINOR% LSS 18 (
                echo Go version is too old. Minimum required is 1.18. | %TEE%
                echo Please upgrade Go from https://go.dev/dl/ | %TEE%
                exit /b 1
            )
        )
    )
    echo Go version check passed. | %TEE%
)

REM Check if Python is installed (only if parameter is passed)
if "%~1"=="with-python" (
    echo Checking Python installation... | %TEE%
    where python >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo Python is not installed. Installing Python... | %TEE%
        echo Please wait, this might take a few minutes... | %TEE%
        
        echo Downloading Python installer... | %TEE%
        powershell -Command "& {Invoke-WebRequest -Uri 'https://www.python.org/ftp/python/3.10.0/python-3.10.0-amd64.exe' -OutFile 'python_installer.exe'}"
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to download Python installer. | %TEE%
            echo Please install Python manually from https://www.python.org/downloads/ | %TEE%
            exit /b 1
        )
        
        echo Installing Python... | %TEE%
        powershell -Command "& {Start-Process -FilePath 'python_installer.exe' -ArgumentList '/quiet', 'InstallAllUsers=1', 'PrependPath=1' -Wait}"
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to install Python. | %TEE%
            echo Please install Python manually from https://www.python.org/downloads/ | %TEE%
            exit /b 1
        )
        
        echo Cleaning up installer... | %TEE%
        del "python_installer.exe"
        
        REM Verify Python installation
        where python >nul 2>&1
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to install Python. Please install it manually from https://www.python.org/downloads/ | %TEE%
            exit /b 1
        )
        echo Python has been installed successfully. | %TEE%
        
        REM Refresh PATH environment variable
        echo Refreshing PATH environment variable... | %TEE%
        for /f "tokens=2*" %%a in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path') do set "PATH=%%b"
        for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v Path') do set "PATH=!PATH!;%%b"
    ) else (
        REM Check Python version
        for /f "tokens=2" %%i in ('python --version 2^>^&1') do set PYTHON_VERSION=%%i
        echo Detected Python version: %PYTHON_VERSION% | %TEE%
        
        REM Simple version check (requires Python 3.8 or higher)
        for /f "tokens=1,2 delims=." %%a in ("%PYTHON_VERSION%") do (
            set MAJOR=%%a
            set MINOR=%%b
        )
        
        if %MAJOR% LSS 3 (
            echo Python version is too old. Minimum required is 3.8. | %TEE%
            echo Please upgrade Python from https://www.python.org/downloads/ | %TEE%
            exit /b 1
        ) else (
            if %MAJOR% EQU 3 (
                if %MINOR% LSS 8 (
                    echo Python version is too old. Minimum required is 3.8. | %TEE%
                    echo Please upgrade Python from https://www.python.org/downloads/ | %TEE%
                    exit /b 1
                )
            )
        )
        echo Python version check passed. | %TEE%
    )
    
    REM Check if pip is installed
    echo Checking pip installation... | %TEE%
    python -m pip --version >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo pip is not installed. Installing pip... | %TEE%
        
        echo Downloading get-pip.py... | %TEE%
        powershell -Command "& {Invoke-WebRequest -Uri 'https://bootstrap.pypa.io/get-pip.py' -OutFile 'get-pip.py'}"
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to download get-pip.py. | %TEE%
            echo Please install pip manually. | %TEE%
            exit /b 1
        )
        
        echo Installing pip... | %TEE%
        python get-pip.py
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to install pip. | %TEE%
            echo Please install pip manually. | %TEE%
            exit /b 1
        }
        
        echo Cleaning up installer... | %TEE%
        del "get-pip.py"
        
        REM Verify pip installation
        python -m pip --version >nul 2>&1
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to install pip. Please install it manually. | %TEE%
            exit /b 1
        )
        echo pip has been installed successfully. | %TEE%
    ) else (
        echo pip is already installed. | %TEE%
    )
)

echo All common dependencies checked and installed successfully. | %TEE%
exit /b 0

endlocal

