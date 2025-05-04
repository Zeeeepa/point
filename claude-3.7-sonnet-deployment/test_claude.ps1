# PowerShell script to test Claude 3.7 Sonnet deployment

Write-Host "`n=== Testing Claude 3.7 Sonnet Deployment ===`n" -ForegroundColor Blue

# Check if the service is running
try {
    $null = Invoke-RestMethod -Uri "http://localhost:8080/v1/models" -Method Get -ErrorAction Stop
    Write-Host "Service is running. Testing Claude 3.7 Sonnet..." -ForegroundColor Green
} catch {
    Write-Host "The chatgpt-adapter service is not running. Please start it first." -ForegroundColor Red
    Write-Host "Run: docker-compose up -d"
    exit 1
}

# Get token from file
if (Test-Path ".cursor_token") {
    $TOKEN = Get-Content ".cursor_token" -Raw
    $TOKEN = $TOKEN.Trim()
} else {
    Write-Host "No token file found. Using the API without authentication." -ForegroundColor Yellow
    $TOKEN = ""
}

# Set headers
$headers = @{
    "Content-Type" = "application/json"
}

if ($TOKEN -ne "") {
    $headers["Authorization"] = "Bearer $TOKEN"
}

# Make the API call
Write-Host "Sending test request to Claude 3.7 Sonnet...`n"

$body = @{
    model = "cursor/claude-3.7-sonnet"
    messages = @(
        @{
            role = "user"
            content = "Hello, are you Claude 3.7 Sonnet? Please confirm your model name and provide a brief greeting."
        }
    )
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/v1/chat/completions" -Method Post -Headers $headers -Body $body
    
    Write-Host "Response received:" -ForegroundColor Green
    $content = $response.choices[0].message.content
    Write-Host $content -ForegroundColor Cyan
    
    if ($content -match "Claude 3.7 Sonnet") {
        Write-Host "`nSuccess! Claude 3.7 Sonnet is working correctly." -ForegroundColor Green
    } else {
        Write-Host "`nThe response doesn't explicitly confirm Claude 3.7 Sonnet. Please check the response content." -ForegroundColor Yellow
    }
} catch {
    Write-Host "Error calling the API: $_" -ForegroundColor Red
}

Write-Host ""

