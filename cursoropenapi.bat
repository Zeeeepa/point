@echo off
echo Starting Cursor OpenAI API compatible endpoint...
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

REM Navigate to the chatgpt-adapter-main directory
echo Checking for chatgpt-adapter-main directory...
if not exist "%~dp0chatgpt-adapter-main" (
    echo Error: chatgpt-adapter-main directory not found.
    echo Please ensure the directory exists in the same location as this batch file.
    pause
    exit /b 1
)
cd "%~dp0chatgpt-adapter-main"

REM Check if dependencies are installed
if not exist "go.mod" (
    echo Initializing Go module...
    go mod init chatgpt-adapter
    echo Installing dependencies...
    go mod download
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to download dependencies.
        pause
        exit /b 1
    )
) else (
    echo Ensuring dependencies are up to date...
    go mod download
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to update dependencies.
        pause
        exit /b 1
    )
)

REM Build and run the adapter
echo Building and starting the Cursor OpenAI API compatible endpoint...
echo.
echo Note: This will create an OpenAI API compatible endpoint at http://localhost:8000
echo You can use this endpoint with any OpenAI API client by setting the base URL to http://localhost:8000
echo.
go run main.go --mode cursor

REM If the program exits, pause to see any error messages
pause
