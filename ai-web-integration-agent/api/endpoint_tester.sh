#!/bin/bash
# Streamlined script for testing Claude, Copilot, and Cursor endpoints

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Function to print colored text
print_color() {
  echo -e "${1}${2}${NC}"
}

# Function to print header
print_header() {
  echo
  echo "================================================================================"
  print_color "${PURPLE}${BOLD}" "  $1"
  echo "================================================================================"
  echo
}

# Function to print status
print_status() {
  if [ "$2" = "success" ]; then
    print_color "${GREEN}" "✓ $1"
  else
    print_color "${RED}" "✗ $1"
  fi
}

# Function to print info
print_info() {
  print_color "${BLUE}" "ℹ $1"
}

# Function to print warning
print_warning() {
  print_color "${YELLOW}" "⚠ $1"
}

# Function to check if a port is in use
is_port_in_use() {
  if command -v nc &> /dev/null; then
    nc -z localhost $1 &> /dev/null
    return $?
  elif command -v lsof &> /dev/null; then
    lsof -i:$1 &> /dev/null
    return $?
  else
    # Fallback to a simple check using /dev/tcp on Linux/macOS
    (echo > /dev/tcp/localhost/$1) &> /dev/null
    return $?
  fi
}

# Function to start the API server
start_server() {
  local port=$1
  local service=$2
  
  # Check if port is already in use
  if is_port_in_use $port; then
    print_warning "Port $port is already in use. This might be another instance of the API server."
    read -p "Do you want to continue anyway? (y/n): " response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
      print_info "Exiting..."
      exit 0
    fi
  fi
  
  print_info "Starting API server for $service on port $port..."
  
  # Determine the correct path to the API server executable
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  server_path="$script_dir/web-integration-api"
  
  # Check if the server executable exists
  if [ ! -f "$server_path" ]; then
    # Try to build it
    print_info "Server executable not found. Attempting to build it..."
    if command -v go &> /dev/null; then
      (cd "$script_dir" && go build -o web-integration-api .)
      if [ $? -eq 0 ]; then
        print_status "Server built successfully" "success"
      else
        print_status "Failed to build server" "failure"
        print_info "Please build the server manually using 'go build -o web-integration-api .'"
        exit 1
      fi
    else
      print_status "Go compiler not found. Please install Go and build the server manually." "failure"
      exit 1
    fi
  fi
  
  # Create a config file for the server
  cat > "$script_dir/config.json" << EOF
{
  "port": $port,
  "host": "0.0.0.0",
  "claude_url": "https://claude.ai/chat",
  "github_copilot_url": "https://github.com/features/copilot",
  "browser_user_data_dir": "~/.browser-agent",
  "screenshot_dir": "./screenshots",
  "log_file": "./api-server.log",
  "headless": false,
  "debug_mode": true
}
EOF
  
  # Start the server
  "$server_path" --config "$script_dir/config.json" --port "$port" > /dev/null 2>&1 &
  server_pid=$!
  
  # Give the server some time to start
  sleep 3
  
  # Check if the server is running
  if ! kill -0 $server_pid 2>/dev/null; then
    print_status "Server failed to start" "failure"
    return 1
  fi
  
  print_status "API server started on port $port" "success"
  echo $server_pid > "$script_dir/.server_pid"
  return 0
}

# Function to stop the API server
stop_server() {
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  pid_file="$script_dir/.server_pid"
  
  if [ -f "$pid_file" ]; then
    server_pid=$(cat "$pid_file")
    print_info "Stopping API server (PID: $server_pid)..."
    
    # Try to terminate gracefully first
    kill -15 $server_pid 2>/dev/null
    
    # Wait for a bit to see if it terminates
    for i in {1..5}; do
      if ! kill -0 $server_pid 2>/dev/null; then
        print_status "API server stopped" "success"
        rm "$pid_file"
        return 0
      fi
      sleep 1
    done
    
    # Force kill if it doesn't terminate
    print_warning "Server not responding to termination signal. Force killing..."
    kill -9 $server_pid 2>/dev/null
    print_status "API server killed" "success"
    rm "$pid_file"
  else
    print_info "No running server found"
  fi
}

# Function to test the API endpoint
test_endpoint() {
  local port=$1
  local service=$2
  local model=""
  local endpoint=""
  local request_data=""
  
  case $service in
    "Claude")
      model="web_claude"
      endpoint="/v1/chat/completions"
      request_data='{
        "model": "web_claude",
        "messages": [
          {"role": "user", "content": "Hello"}
        ],
        "temperature": 0.7,
        "stream": false
      }'
      ;;
    "GitHub Copilot")
      model="web_copilot"
      endpoint="/v1/completions"
      request_data='{
        "model": "web_copilot",
        "prompt": "// Say hello\nfunction greet() {",
        "temperature": 0.7,
        "stream": false
      }'
      ;;
    "Cursor")
      model="cursor/claude-3.7-sonnet"
      endpoint="/v1/chat/completions"
      request_data='{
        "model": "cursor/claude-3.7-sonnet",
        "messages": [
          {"role": "user", "content": "Hello"}
        ],
        "temperature": 0.7,
        "stream": false
      }'
      ;;
  esac
  
  print_info "Testing $service endpoint..."
  
  # Check if curl is available
  if ! command -v curl &> /dev/null; then
    print_status "curl command not found. Please install curl to test the endpoint." "failure"
    return 1
  fi
  
  # Test the endpoint
  response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -d "$request_data" \
    "http://localhost:$port$endpoint")
  
  # Extract status code and response body
  status_code=$(echo "$response" | tail -n1)
  response_body=$(echo "$response" | sed '$d')
  
  if [ "$status_code" -eq 200 ]; then
    print_status "$service endpoint is working!" "success"
    print_info "Response:"
    
    # Pretty print the response if jq is available
    if command -v jq &> /dev/null; then
      echo "$response_body" | jq .
    else
      echo "$response_body"
    fi
    return 0
  else
    print_status "Endpoint test failed with status code: $status_code" "failure"
    print_info "Response:"
    echo "$response_body"
    return 1
  fi
}

# Function to clear the screen
clear_screen() {
  clear
}

# Main menu
main_menu() {
  clear_screen
  print_header "AI Web Integration API Tester"
  
  echo "Select an AI service to test:"
  echo "1. Claude"
  echo "2. GitHub Copilot"
  echo "3. Cursor"
  echo "4. Exit"
  
  while true; do
    read -p $'\nEnter your choice (1-4): ' choice
    case $choice in
      1)
        service="Claude"
        break
        ;;
      2)
        service="GitHub Copilot"
        break
        ;;
      3)
        service="Cursor"
        break
        ;;
      4)
        print_info "Exiting..."
        exit 0
        ;;
      *)
        print_warning "Invalid choice. Please try again."
        ;;
    esac
  done
  
  # Get port number
  while true; do
    read -p $'\nEnter port number (default: 8080): ' port_input
    if [ -z "$port_input" ]; then
      port=8080
      break
    fi
    
    if [[ "$port_input" =~ ^[0-9]+$ ]] && [ "$port_input" -ge 1024 ] && [ "$port_input" -le 65535 ]; then
      port=$port_input
      break
    else
      print_warning "Port must be a number between 1024 and 65535."
    fi
  done
  
  echo "$service:$port"
}

# Main function
main() {
  # Trap Ctrl+C to clean up
  trap 'print_info "Interrupted by user. Cleaning up..."; stop_server; exit 1' INT
  
  while true; do
    # Show menu and get choices
    IFS=':' read -r service port <<< "$(main_menu)"
    
    # Start server
    if ! start_server "$port" "$service"; then
      read -p $'\nPress Enter to continue...' dummy
      continue
    fi
    
    # Test endpoint
    test_endpoint "$port" "$service"
    
    # Ask what to do next
    echo -e "\nWhat would you like to do next?"
    echo "1. Test another service"
    echo "2. Exit"
    
    while true; do
      read -p $'\nEnter your choice (1-2): ' next_choice
      case $next_choice in
        1|2)
          break
          ;;
        *)
          print_warning "Invalid choice. Please try again."
          ;;
      esac
    done
    
    # Stop the current server
    stop_server
    
    if [ "$next_choice" -eq 2 ]; then
      print_info "Exiting..."
      break
    fi
  done
}

# Run the main function
main

