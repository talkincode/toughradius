#!/bin/bash

# ToughRADIUS admin password reset script
# Usage: ./reset-admin-password.sh [new password]

set -e

CONFIG_FILE="toughradius.yml"
NEW_PASSWORD="${1:-toughradius}"

echo "========================================"
echo "ToughRADIUS Admin Password Reset Tool"
echo "========================================"
echo ""

# Check the configuration file
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Configuration file $CONFIG_FILE not found"
    echo "Please run this script in the ToughRADIUS root directory"
    exit 1
fi

# Detect the database type
DB_TYPE=$(grep -A 5 "^database:" "$CONFIG_FILE" | grep "type:" | awk '{print $2}' | tr -d '"' || echo "postgres")

# Decide whether to enable CGO based on the database type
if [ "$DB_TYPE" = "sqlite" ]; then
    echo "Detected SQLite database, enabling CGO..."
    export CGO_ENABLED=1
else
    echo "Detected $DB_TYPE database, using static compilation..."
    export CGO_ENABLED=0
fi

# Build the password reset tool
echo "Building password reset tool..."
cd cmd/reset-password
go build -o ../../reset-password .
cd ../..

# Run the password reset tool
echo "Resetting admin password..."
./reset-password -c "$CONFIG_FILE" -u admin -p "$NEW_PASSWORD"

# Clean up
rm -f reset-password

echo ""
echo "========================================"
echo "Password reset completed!"
echo "========================================"
echo ""
echo "Login information:"
echo "  Username: admin"
echo "  Password: $NEW_PASSWORD"
echo "  Access URL: http://localhost:1816"
echo ""
