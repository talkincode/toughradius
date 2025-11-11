#!/bin/bash

###############################################################################
# ToughRADIUS database initialization script
#
# Purpose: quickly initialize the ToughRADIUS database (PostgreSQL or SQLite supported)
#
# Usage:
#   1. PostgreSQL: ./scripts/init-db.sh postgres
#   2. SQLite:     ./scripts/init-db.sh sqlite
#   3. Use a config file: ./scripts/init-db.sh
#
# Examples:
#   ./scripts/init-db.sh sqlite
#   ./scripts/init-db.sh postgres
###############################################################################

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Determine the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}ToughRADIUS Database Initialization${NC}"
echo -e "${GREEN}================================${NC}"
echo ""

# Check arguments
DB_TYPE="${1:-}"

if [ -z "$DB_TYPE" ]; then
    echo -e "${YELLOW}No database type specified, using configuration file settings${NC}"
    echo ""
    cd "$PROJECT_ROOT"
    
    # Check if the tool has been compiled
    if [ ! -f "./toughradius" ]; then
        echo -e "${YELLOW}Binary not found, starting compilation...${NC}"
        go build -o toughradius
        echo -e "${GREEN}✓ Compilation completed${NC}"
    fi
    
    echo -e "${YELLOW}Initializing database...${NC}"
    ./toughradius -initdb
    
    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}Database initialization completed!${NC}"
    echo -e "${GREEN}================================${NC}"
    echo ""
    echo -e "${GREEN}Default admin account:${NC}"
    echo -e "  Username: ${YELLOW}admin${NC}"
    echo -e "  Password: ${YELLOW}toughradius${NC}"
    echo ""
    echo -e "${GREEN}API user account:${NC}"
    echo -e "  Username: ${YELLOW}apiuser${NC}"
    echo -e "  Password: ${YELLOW}Api_189${NC}"
    echo ""
    
    exit 0
fi

# Parameterized initialization
case "$DB_TYPE" in
    sqlite)
        echo -e "${GREEN}Using SQLite database${NC}"
        echo ""
        
        # Create a temporary configuration file
        TEMP_CONFIG="/tmp/toughradius-sqlite.yml"
        cat > "$TEMP_CONFIG" <<EOF
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: /var/toughradius
  debug: true

web:
  host: 0.0.0.0
  port: 1816
  tls_port: 1817
  secret: 9b6de5cc-0731-1203-xxtt-0f568ac9da37

database:
  type: sqlite
  name: toughradius.db
  max_conn: 100
  idle_conn: 10
  debug: false

freeradius:
  enabled: true
  host: 0.0.0.0
  port: 1818
  debug: true

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  radsec_worker: 100
  debug: true

logger:
  mode: development
  file_enable: true
  filename: /var/toughradius/toughradius.log
EOF
        
        echo -e "${YELLOW}SQLite database configuration:${NC}"
        echo -e "  Database file: ${YELLOW}/var/toughradius/data/toughradius.db${NC}"
        echo ""
        
        cd "$PROJECT_ROOT"
        
        # Check if the tool has been compiled
        if [ ! -f "./toughradius" ]; then
            echo -e "${YELLOW}Binary not found, starting compilation...${NC}"
            go build -o toughradius
            echo -e "${GREEN}✓ Compilation completed${NC}"
        fi
        
        # Create the data directory
        sudo mkdir -p /var/toughradius/data
        sudo chown -R $(whoami) /var/toughradius
        
        echo -e "${YELLOW}Initializing SQLite database...${NC}"
        ./toughradius -c "$TEMP_CONFIG" -initdb
        
        # Clean up temporary configuration
        rm -f "$TEMP_CONFIG"
        
        echo ""
        echo -e "${GREEN}================================${NC}"
        echo -e "${GREEN}SQLite database initialization completed!${NC}"
        echo -e "${GREEN}================================${NC}"
        echo ""
        echo -e "${GREEN}Database location:${NC}"
        echo -e "  ${YELLOW}/var/toughradius/data/toughradius.db${NC}"
        echo ""
        echo -e "${GREEN}Start command (using SQLite):${NC}"
        echo -e "  ${YELLOW}./toughradius -c $TEMP_CONFIG${NC}"
        echo ""
        echo -e "${GREEN}Or modify toughradius.yml configuration file:${NC}"
        cat <<EOF
  ${YELLOW}database:
    type: sqlite
    name: toughradius.db${NC}
EOF
        echo ""
    ;;
    
    postgres|postgresql)
        echo -e "${GREEN}Using PostgreSQL database${NC}"
        echo ""
        
        # Load the configuration
        read -p "PostgreSQL host [127.0.0.1]: " PG_HOST
        PG_HOST=${PG_HOST:-127.0.0.1}
        
        read -p "PostgreSQL port [5432]: " PG_PORT
        PG_PORT=${PG_PORT:-5432}
        
        read -p "Database name [toughradius]: " PG_DB
        PG_DB=${PG_DB:-toughradius}
        
        read -p "Database user [postgres]: " PG_USER
        PG_USER=${PG_USER:-postgres}
        
        read -sp "Database password: " PG_PASS
        echo ""
        
        # Create a temporary configuration file
        TEMP_CONFIG="/tmp/toughradius-postgres.yml"
        cat > "$TEMP_CONFIG" <<EOF
system:
  appid: ToughRADIUS
  location: Asia/Shanghai
  workdir: /var/toughradius
  debug: true

web:
  host: 0.0.0.0
  port: 1816
  tls_port: 1817
  secret: 9b6de5cc-0731-1203-xxtt-0f568ac9da37

database:
  type: postgres
  host: $PG_HOST
  port: $PG_PORT
  name: $PG_DB
  user: $PG_USER
  passwd: $PG_PASS
  max_conn: 100
  idle_conn: 10
  debug: false

freeradius:
  enabled: true
  host: 0.0.0.0
  port: 1818
  debug: true

radiusd:
  enabled: true
  host: 0.0.0.0
  auth_port: 1812
  acct_port: 1813
  radsec_port: 2083
  radsec_worker: 100
  debug: true

logger:
  mode: development
  file_enable: true
  filename: /var/toughradius/toughradius.log
EOF
        
        echo ""
        echo -e "${YELLOW}PostgreSQL database configuration:${NC}"
        echo -e "  Host: ${YELLOW}$PG_HOST:$PG_PORT${NC}"
        echo -e "  Database: ${YELLOW}$PG_DB${NC}"
        echo -e "  User: ${YELLOW}$PG_USER${NC}"
        echo ""
        
        cd "$PROJECT_ROOT"
        
        # Check if the tool has been compiled
        if [ ! -f "./toughradius" ]; then
            echo -e "${YELLOW}Binary not found, starting compilation...${NC}"
            go build -o toughradius
            echo -e "${GREEN}✓ Compilation completed${NC}"
        fi
        
        echo -e "${YELLOW}Initializing PostgreSQL database...${NC}"
        ./toughradius -c "$TEMP_CONFIG" -initdb
        
        # Clean up temporary configuration
        rm -f "$TEMP_CONFIG"
        
        echo ""
        echo -e "${GREEN}================================${NC}"
        echo -e "${GREEN}PostgreSQL database initialization completed!${NC}"
        echo -e "${GREEN}================================${NC}"
        echo ""
    ;;
    
    *)
        echo -e "${RED}Error: Unsupported database type '$DB_TYPE'${NC}"
        echo ""
        echo -e "${YELLOW}Supported database types:${NC}"
        echo "  - sqlite"
        echo "  - postgres (or postgresql)"
        echo ""
        echo -e "${YELLOW}Usage examples:${NC}"
        echo "  ./scripts/init-db.sh sqlite"
        echo "  ./scripts/init-db.sh postgres"
        echo ""
        exit 1
    ;;
esac

echo -e "${GREEN}Default admin account:${NC}"
echo -e "  Username: ${YELLOW}admin${NC}"
echo -e "  Password: ${YELLOW}toughradius${NC}"
echo ""
echo -e "${GREEN}API user account:${NC}"
echo -e "  Username: ${YELLOW}apiuser${NC}"
echo -e "  Password: ${YELLOW}Api_189${NC}"
echo ""
echo -e "${GREEN}Web admin interface:${NC}"
echo -e "  ${YELLOW}http://localhost:1816/admin${NC}"
echo ""
