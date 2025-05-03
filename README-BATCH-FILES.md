# OpenAI API Compatible Endpoints - Batch Files

This directory contains batch files for setting up and running OpenAI API compatible endpoints for various AI services on Windows.

## Available Endpoints

1. **Cursor OpenAI API** - Connect to Cursor AI
2. **Web Claude OpenAI API** - Connect to Claude AI
3. **Web Copilot OpenAI API** - Connect to GitHub Copilot

## Prerequisites

- Windows operating system
- Internet connection
- Administrator privileges (for installing dependencies)

## Batch Files

### Main Endpoint Files

- **`cursoropenapi.bat`** - Starts an OpenAI API compatible endpoint connected to Cursor (v1.1.0 with enhanced logging and fallback mechanisms)
- **`webchatclaude.bat`** - Starts an OpenAI API compatible endpoint connected to Claude
- **`webchatcopilot.bat`** - Starts an OpenAI API compatible endpoint connected to GitHub Copilot

### Helper Scripts

- **`install-common.bat`** - Handles installation of common dependencies (Go, Python)
- **`install-playwright.bat`** - Handles installation of Playwright and browser dependencies

## Usage

1. Make sure you have the required directories:
   - `chatgpt-adapter-main` - For Cursor endpoint
   - `ai-web-integration-agent` - For Claude and Copilot endpoints

2. Run the desired batch file by double-clicking it or running it from the command line:
   ```
   cursoropenapi.bat
   ```
   or
   ```
   webchatclaude.bat
   ```
   or
   ```
   webchatcopilot.bat
   ```

3. The batch file will:
   - Check and install required dependencies (Go, Python)
   - Install necessary packages and libraries
   - Create default configuration files if needed
   - Start the appropriate service

4. For Claude and GitHub Copilot endpoints, you will need to log in to the respective services when the browser opens.

## Configuration

- Default configuration files are created in the `config` directory
- You can modify these files to change settings like:
  - URLs for the services
  - Browser data directory
  - Headless mode
  - Debug settings

## Troubleshooting

If you encounter issues:

1. Check the log files in the `logs` directory for detailed error information
2. Ensure you have a stable internet connection
3. Make sure you have the required directories in the correct location
4. Try running the batch files as administrator
5. Check that you have sufficient disk space
6. If a port is already in use, the script will automatically try alternative ports

## Advanced Features (cursoropenapi.bat v1.1.0)

- **Automatic Logging**: All operations are logged to timestamped files in the `logs` directory
- **Fallback Mechanisms**: 
  - Automatic retry for failed dependency installations
  - Alternative port selection if the default port is in use
  - Multiple download methods if the repository is missing
- **Error Recovery**: Detailed error messages and recovery steps for common issues
- **Port Management**: Automatic detection of port conflicts and selection of alternative ports

## Notes

- The first run may take longer as it installs dependencies
- Browser binaries will be downloaded during the Playwright installation
- You will need to authenticate with the respective services (Claude, GitHub)
- Log files are stored in the `logs` directory with timestamps for easy troubleshooting
