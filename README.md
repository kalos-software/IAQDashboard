# Sensor Dashboard System

A complete environmental monitoring solution with a Go-based API server and Svelte dashboard frontend for visualizing sensor data.

## Overview

This system consists of two main components:
- **Sensor API**: A Go backend that provides RESTful endpoints for storing and retrieving sensor data
- **Sensor Dashboard**: A Svelte-based web frontend for visualizing sensor data with interactive charts

## Requirements

- Raspberry Pi (Raspberry Pi 4 or newer recommended)
- MySQL database
- Bun (for building the Svelte app)
- Go (for building the API)
- Apache web server

## Installation

### 1. Install Go
```bash
# Download the appropriate version for ARM
wget https://go.dev/dl/go1.21.4.linux-arm64.tar.gz

# Extract it
sudo tar -C /usr/local -xzf go1.21.4.linux-arm64.tar.gz

# Add Go to path
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile

# Verify installation
go version
```

### 2. Install Bun

```bash
# Install dependencies
sudo apt-get install -y unzip

# Install Bun
curl -fsSL https://bun.sh/install | bash

# Source the updated profile to use Bun
source ~/.bashrc
# Or
source ~/.zshrc

# Verify installation
bun --version
```

### 3. Build the Svelte Dashboard

```bash
# Navigate to the dashboard directory
cd sensor-dashboard

# Install dependencies
bun install

# Build the production bundle
bun run build

# Move the build to Apache's web directory
sudo rm -rf /var/www/html/*
sudo cp -r build/* /var/www/html/
```

### 4. Build the Go API

```bash
# Navigate to the API directory
cd sensor-api

# Get dependencies
go mod download

# Build the binary
go build -o sensor-api main.go

# Make it executable
chmod +x sensor-api

# Move to a suitable location
sudo mv sensor-api /usr/local/bin/
```

### 5. Create a Systemd Service for the API

Create a systemd service file to run the API on boot:

```bash
sudo nano /etc/systemd/system/sensor-api.service
```

Add the following content:

```
[Unit]
Description=Sensor API Server
After=network.target mysql.service

[Service]
ExecStart=/usr/local/bin/sensor-api
Restart=always
User=pi
Group=pi
Environment=PORT=8080
# Add any other environment variables here if needed

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable sensor-api
sudo systemctl start sensor-api
```

## Configuration

### Environment Variables

The Go API uses the following environment variables for database connection:

- Host: $MYSQL_HOST
- User: $MYSQL_USER
- Password: $MYSQL_PASS
- Database: $MYSQL_DB
- Port: $MYSQL_PORT

If you need to change these, modify the `dbConfig` struct in the `main.go` file and rebuild.

### Dashboard Configuration

The dashboard is configured to connect to the API at the same host. If you need to change the API URL, modify `src/lib/config.ts` in the Svelte app.

## Usage

After installation:

1. The API will automatically start on boot and listen on port 8080
2. Access the dashboard by navigating to `http://raspberry-pi-ip` in your browser
3. Use the dashboard to view sensor data with various filtering options

## Troubleshooting

### Checking Service Status

```bash
# Check if the API service is running
sudo systemctl status sensor-api

# View API logs
sudo journalctl -u sensor-api -f
```

### Database Connection Issues

```bash
# Check MySQL is running
sudo systemctl status mysql

# Test database connection
mysql -u data -p -h localhost buildingData
```

### Web Server Issues

```bash
# Check Apache status
sudo systemctl status apache2

# Check Apache error logs
sudo tail -f /var/log/apache2/error.log
```

## Maintenance

### Updating the Dashboard

To update the dashboard:

```bash
cd sensor-dashboard
git pull
bun install
bun run build
sudo cp -r build/* /var/www/html/
```

### Updating the API

To update the API:

```bash
cd sensor-api
git pull
go build -o sensor-api main.go
sudo systemctl stop sensor-api
sudo mv sensor-api /usr/local/bin/
sudo systemctl start sensor-api
```

### Backing Up the Database

```bash
# Create a backup
mysqldump -u data -p buildingData > buildingData_backup.sql

# Restore from backup
mysql -u data -p buildingData < buildingData_backup.sql
```