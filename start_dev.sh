#!/bin/bash

# Configuration
CONFIG_FILE="toughradius.dev.yml"
BACKEND_PORT=1816
FRONTEND_PORT=3000

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting ToughRadius Development Environment...${NC}"

# Clean up stale processes
echo -e "${GREEN}Cleaning up stale processes...${NC}"
fuser -k ${BACKEND_PORT}/tcp 2>/dev/null
fuser -k ${FRONTEND_PORT}/tcp 2>/dev/null
pkill -9 -f "go run main.go" 2>/dev/null
pkill -9 -f "vite" 2>/dev/null
sleep 2

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed or not in PATH.${NC}"
    # Try to add default user go path
    export PATH=$HOME/go/go/bin:$PATH
    if ! command -v go &> /dev/null; then
         echo -e "${RED}Still could not find Go. Please install Go 1.24+.${NC}"
         exit 1
    else
         echo -e "${GREEN}Found Go after updating PATH.${NC}"
    fi
fi

# Check Configuration
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}Configuration file $CONFIG_FILE not found.${NC}"
    echo "Creating default $CONFIG_FILE..."
    cat > $CONFIG_FILE <<EOF
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: ./rundata

database:
  type: sqlite
  name: toughradius.db

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083

web:
  host: 0.0.0.0
  port: ${BACKEND_PORT}
EOF
fi

# Frontend Setup
echo -e "${GREEN}Setting up Frontend...${NC}"
cd web
if [ ! -d "node_modules" ]; then
    npm install
fi

# Build is required for go:embed if dist doesn't exist
if [ ! -d "dist" ]; then
    echo "Building frontend assets for embedding..."
    npm run build
fi
cd ..

# Database Initialization
echo -e "${GREEN}Initializing Database...${NC}"
go run main.go -initdb -c $CONFIG_FILE

# Start Servers
echo -e "${GREEN}Starting Backend and Frontend...${NC}"

# Trap Ctrl+C to kill both processes
trap 'kill $(jobs -p)' EXIT

# Start Backend in background
nohup go run main.go -c $CONFIG_FILE > backend.log 2>&1 &
BACKEND_PID=$!

# Wait for backend to be ready
echo -e "${GREEN}Waiting for backend to be ready on port ${BACKEND_PORT}...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:${BACKEND_PORT}/api/v1/system/settings &> /dev/null; then
        echo -e "${GREEN}Backend is ready!${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}Backend failed to start after 30 seconds.${NC}"
        tail -n 20 backend.log
        exit 1
    fi
    sleep 1
done

# Start Frontend
echo -e "${GREEN}Starting Frontend on port ${FRONTEND_PORT}...${NC}"
cd web
npm run dev
