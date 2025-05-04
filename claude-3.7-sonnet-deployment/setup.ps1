# PowerShell script to set up Claude 3.7 Sonnet deployment

Write-Host "`n=== Claude 3.7 Sonnet Deployment Setup ===`n" -ForegroundColor Blue
Write-Host "This script will help you set up the chatgpt-adapter for Claude 3.7 Sonnet via Cursor" -ForegroundColor Yellow
Write-Host ""

# Check if Docker is installed
try {
    $null = Get-Command docker -ErrorAction Stop
} catch {
    Write-Host "Docker is not installed. Please install Docker first." -ForegroundColor Red
    Write-Host "Visit https://docs.docker.com/get-docker/ for installation instructions."
    exit 1
}

# Check if Docker Compose is installed
try {
    $null = Get-Command docker-compose -ErrorAction Stop
} catch {
    Write-Host "Docker Compose is not installed. Please install Docker Compose first." -ForegroundColor Red
    Write-Host "Visit https://docs.docker.com/compose/install/ for installation instructions."
    exit 1
}

# Check if Docker is running
try {
    $null = docker info -ErrorAction Stop
} catch {
    Write-Host "Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

# Get Cursor session token
Write-Host "You need a valid Cursor session token (WorkosCursorSessionToken) to use Claude 3.7 Sonnet." -ForegroundColor Yellow
Write-Host "You can get this token by logging into Cursor (https://www.cursor.com) and extracting it from your browser cookies."
Write-Host ""
Write-Host "To get your token:" -ForegroundColor Blue
Write-Host "1. Log in to Cursor (https://www.cursor.com)"
Write-Host "2. Open your browser's developer tools (F12 or right-click > Inspect)"
Write-Host "3. Go to the Application/Storage tab"
Write-Host "4. Find Cookies > https://www.cursor.com"
Write-Host "5. Copy the value of the 'WorkosCursorSessionToken' cookie"
Write-Host ""

$CURSOR_TOKEN = Read-Host "Enter your Cursor session token (WorkosCursorSessionToken)"

if ([string]::IsNullOrWhiteSpace($CURSOR_TOKEN)) {
    Write-Host "No token provided. Setup cannot continue." -ForegroundColor Red
    exit 1
}

# Create a token file
Write-Host "Saving token..." -ForegroundColor Green
$CURSOR_TOKEN | Out-File -FilePath ".cursor_token" -NoNewline

# Start the service
Write-Host "Starting chatgpt-adapter with Claude 3.7 Sonnet support..." -ForegroundColor Green
docker-compose up -d

Write-Host ""
Write-Host "=== Setup Complete! ===" -ForegroundColor Green
Write-Host "The chatgpt-adapter is now running at http://localhost:8080" -ForegroundColor Blue
Write-Host ""
Write-Host "Available Claude 3.7 Models:" -ForegroundColor Yellow
Write-Host "- claude-3.7-sonnet"
Write-Host "- claude-3.7-sonnet-max"
Write-Host "- claude-3.7-sonnet-thinking"
Write-Host "- claude-3.7-sonnet-thinking-max"
Write-Host ""
Write-Host "To test the deployment, run:" -ForegroundColor Blue
Write-Host ".\test_claude.ps1"
Write-Host ""
Write-Host "To stop the service:" -ForegroundColor Yellow
Write-Host "docker-compose down"
Write-Host ""

