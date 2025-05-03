@echo off
REM Common installation script for dependencies
REM This script handles installation of Go and Python dependencies

REM Check if Go is installed and has minimum version
echo Checking Go installation...
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo Go is not installed. Installing Go...
    echo Please wait, this might take a few minutes...
    powershell -Command "& {Invoke-WebRequest -Uri 'https://go.dev/dl/go1.21.0.windows-amd64.msi' -OutFile 'go_installer.msi'; Start-Process -FilePath 'msiexec.exe' -ArgumentList '/i', 'go_installer.msi', '/quiet' -Wait; Remove-Item 'go_installer.msi'}"
    
    REM Verify Go installation
    where go >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to install Go. Please install it manually from https://go.dev/dl/
        exit /b 1
    )
    echo Go has been installed successfully.
) else (
    REM Check Go version
    for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
    echo Detected Go version: %GO_VERSION%
    REM Remove 'go' prefix from version string
    set GO_VERSION=%GO_VERSION:go=%
    
    REM Simple version check (requires Go 1.18 or higher)
    for /f "tokens=1,2 delims=." %%a in ("%GO_VERSION%") do (
        set MAJOR=%%a
        set MINOR=%%b
    )
    
    if %MAJOR% LSS 1 (
        echo Go version is too old. Minimum required is 1.18.
        echo Please upgrade Go from https://go.dev/dl/
        exit /b 1
    ) else (
        if %MAJOR% EQU 1 (
            if %MINOR% LSS 18 (
                echo Go version is too old. Minimum required is 1.18.
                echo Please upgrade Go from https://go.dev/dl/
                exit /b 1
            )
        )
    )
    echo Go version check passed.
)

REM Check if Python is installed (only if parameter is passed)
if "%~1"=="with-python" (
    echo Checking Python installation...
    where python >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo Python is not installed. Installing Python...
        echo Please wait, this might take a few minutes...
        powershell -Command "& {Invoke-WebRequest -Uri 'https://www.python.org/ftp/python/3.10.0/python-3.10.0-amd64.exe' -OutFile 'python_installer.exe'; Start-Process -FilePath 'python_installer.exe' -ArgumentList '/quiet', 'InstallAllUsers=1', 'PrependPath=1' -Wait; Remove-Item 'python_installer.exe'}"
        
        REM Verify Python installation
        where python >nul 2>&1
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to install Python. Please install it manually from https://www.python.org/downloads/
            exit /b 1
        )
        echo Python has been installed successfully.
    ) else (
        REM Check Python version
        for /f "tokens=2" %%i in ('python --version 2^>^&1') do set PYTHON_VERSION=%%i
        echo Detected Python version: %PYTHON_VERSION%
        
        REM Simple version check (requires Python 3.8 or higher)
        for /f "tokens=1,2 delims=." %%a in ("%PYTHON_VERSION%") do (
            set MAJOR=%%a
            set MINOR=%%b
        )
        
        if %MAJOR% LSS 3 (
            echo Python version is too old. Minimum required is 3.8.
            echo Please upgrade Python from https://www.python.org/downloads/
            exit /b 1
        ) else (
            if %MAJOR% EQU 3 (
                if %MINOR% LSS 8 (
                    echo Python version is too old. Minimum required is 3.8.
                    echo Please upgrade Python from https://www.python.org/downloads/
                    exit /b 1
                )
            )
        )
        echo Python version check passed.
    )
    
    REM Check if pip is installed
    echo Checking pip installation...
    python -m pip --version >nul 2>&1
    if %ERRORLEVEL% NEQ 0 (
        echo pip is not installed. Installing pip...
        powershell -Command "& {Invoke-WebRequest -Uri 'https://bootstrap.pypa.io/get-pip.py' -OutFile 'get-pip.py'; python get-pip.py; Remove-Item 'get-pip.py'}"
        
        REM Verify pip installation
        python -m pip --version >nul 2>&1
        if %ERRORLEVEL% NEQ 0 (
            echo Failed to install pip. Please install it manually.
            exit /b 1
        )
        echo pip has been installed successfully.
    ) else (
        echo pip is already installed.
    )
)

echo All common dependencies checked and installed successfully.
exit /b 0

