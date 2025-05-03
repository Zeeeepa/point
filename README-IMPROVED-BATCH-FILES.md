# Improved OpenAI API Compatible Endpoints - Batch Files

This directory contains improved batch files for setting up and running OpenAI API compatible endpoints for various AI services on Windows. These improved versions include better error handling, logging, and debugging capabilities.

## Available Endpoints

1. **Cursor OpenAI API** - Connect to Cursor AI
2. **Web Claude OpenAI API** - Connect to Claude AI
3. **Web Copilot OpenAI API** - Connect to GitHub Copilot

## Improvements Over Original Batch Files

- **Comprehensive Logging**: All output is logged to date-stamped log files in the `logs` directory
- **Verbose Error Handling**: Detailed error messages with exit codes
- **Pause Before Execution**: Prevents accidental closing by requiring user confirmation
- **Tee Output**: Shows output in console while also logging to file
- **Debug Wrapper**: Optional wrapper script for advanced debugging

## Batch Files

### Main Endpoint Files

- **`improved-cursoropenapi.bat`** - Improved version of the Cursor OpenAI API endpoint
- **`improved-webchatclaude.bat`** - Improved version of the Claude OpenAI API endpoint
- **`improved-webchatcopilot.bat`** - Improved version of the GitHub Copilot OpenAI API endpoint

### Helper Scripts

- **`improved-install-common.bat`** - Improved version of the common dependencies installer
- **`improved-install-playwright.bat`** - Improved version of the Playwright installer
- **`debug-wrapper.bat`** - Debug wrapper that can be used with any of the original batch files

## Usage

### Using the Improved Batch Files

Simply run the improved batch files directly:

```
improved-cursoropenapi.bat
```

or

```
improved-webchatclaude.bat
```

or

```
improved-webchatcopilot.bat
```

### Using the Debug Wrapper

The debug wrapper can be used with any of the original batch files to provide better logging and error handling:

```
debug-wrapper.bat cursoropenapi.bat
```

or

```
debug-wrapper.bat webchatclaude.bat
```

or

```
debug-wrapper.bat webchatcopilot.bat
```

## Log Files

All log files are stored in the `logs` directory with timestamps in the filename:

- `cursoropenapi_YYYYMMDD-HHMMSS.log`
- `webchatclaude_YYYYMMDD-HHMMSS.log`
- `webchatcopilot_YYYYMMDD-HHMMSS.log`
- `install-common_YYYYMMDD-HHMMSS.log`
- `install-playwright_YYYYMMDD-HHMMSS.log`

## Troubleshooting

If you encounter issues:

1. Check the log files in the `logs` directory for detailed error messages
2. Use the debug wrapper for more comprehensive logging
3. Ensure you have administrator privileges when running the batch files
4. Verify that all required directories exist:
   - `chatgpt-adapter-main` - For Cursor endpoint
   - `ai-web-integration-agent` - For Claude and Copilot endpoints
5. Check your internet connection
6. Ensure you have sufficient disk space

## Common Issues and Solutions

1. **"Command not found" errors**: Make sure Go and Python are properly installed and in your PATH
2. **Permission errors**: Run the batch files as administrator
3. **Missing dependencies**: The batch files will attempt to install dependencies, but may require manual intervention
4. **Browser launch failures**: Make sure you have a compatible browser installed
5. **Authentication issues**: You will need to log in to the respective services when prompted

## Notes

- The first run may take longer as it installs dependencies
- Browser binaries will be downloaded during the Playwright installation
- You will need to authenticate with the respective services (Claude, GitHub)
- All batch files will pause at the end to allow you to read any error messages

