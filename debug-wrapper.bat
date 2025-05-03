@echo off
setlocal enabledelayedexpansion

REM Create logs directory if it doesn't exist
if not exist "%~dp0logs" mkdir "%~dp0logs"

REM Get the current date and time for the log filename
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "LOGDATE=%dt:~0,8%-%dt:~8,6%"
set "LOGFILE=%~dp0logs\%~1_%LOGDATE%.log"

echo ===================================================
echo Debug Wrapper for %~1
echo Log file: %LOGFILE%
echo ===================================================
echo.
echo Press any key to start execution...
pause > nul

REM Run the specified batch file and redirect all output to the log file
echo Starting execution of %~1 at %TIME% > "%LOGFILE%"
echo Environment information: >> "%LOGFILE%"
echo OS Version: >> "%LOGFILE%"
ver >> "%LOGFILE%"
echo System PATH: >> "%LOGFILE%"
echo %PATH% >> "%LOGFILE%"
echo. >> "%LOGFILE%"
echo ===================================================== >> "%LOGFILE%"
echo. >> "%LOGFILE%"

REM Run the batch file with output redirection
call "%~dp0%~1" >> "%LOGFILE%" 2>&1
set EXIT_CODE=%ERRORLEVEL%

echo. >> "%LOGFILE%"
echo ===================================================== >> "%LOGFILE%"
echo Execution of %~1 completed at %TIME% with exit code %EXIT_CODE% >> "%LOGFILE%"

REM Display the results
echo.
echo Execution of %~1 completed with exit code %EXIT_CODE%
echo Log file created at: %LOGFILE%
echo.

REM If there was an error, show the last few lines of the log
if %EXIT_CODE% NEQ 0 (
    echo Last 20 lines of the log file:
    echo -----------------------------------------------------
    powershell -Command "Get-Content -Path '%LOGFILE%' -Tail 20"
    echo -----------------------------------------------------
)

echo.
echo Press any key to open the log file, or CTRL+C to exit...
pause > nul

REM Open the log file
start notepad "%LOGFILE%"

endlocal

