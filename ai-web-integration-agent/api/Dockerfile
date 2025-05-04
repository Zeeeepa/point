FROM golang:1.20-bullseye

# Install Chrome dependencies
RUN apt-get update && apt-get install -y \
    wget \
    gnupg \
    ca-certificates \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libatspi2.0-0 \
    libcups2 \
    libdbus-1-3 \
    libdrm2 \
    libgbm1 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libwayland-client0 \
    libxcomposite1 \
    libxdamage1 \
    libxfixes3 \
    libxkbcommon0 \
    libxrandr2 \
    xdg-utils \
    --no-install-recommends

# Install Chrome
RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - \
    && echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list \
    && apt-get update \
    && apt-get install -y google-chrome-stable --no-install-recommends

# Install Node.js and npm for Playwright
RUN curl -fsSL https://deb.nodesource.com/setup_16.x | bash - \
    && apt-get install -y nodejs

# Install Playwright
RUN npm install -g playwright@1.32.0 \
    && npx playwright install chromium

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o web-integration-api ./api

# Create directories for data
RUN mkdir -p /data/browser-data /data/screenshots /data/logs

# Set environment variables
ENV BROWSER_USER_DATA_DIR=/data/browser-data
ENV SCREENSHOT_DIR=/data/screenshots
ENV LOG_FILE=/data/logs/api-server.log

# Expose port
EXPOSE 8080

# Run the application
CMD ["./web-integration-api", "--config", "api/config.json"]

