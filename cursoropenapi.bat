@echo off
echo Starting Cursor OpenAI API compatible endpoint...
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

REM Navigate to the chatgpt-adapter-main directory
cd "%~dp0chatgpt-adapter-main"

REM Check if dependencies are installed
if not exist "go.sum" (
    echo Installing dependencies...
    go mod download
    if %ERRORLEVEL% NEQ 0 (
        echo Failed to download dependencies.
        pause
        exit /b 1
    )
)

REM Build and run the adapter
echo Building and starting the Cursor OpenAI API compatible endpoint...
go run main.go --mode cursor

REM If the program exits, pause to see any error messages
pause

