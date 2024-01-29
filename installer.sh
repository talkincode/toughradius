#!/bin/bash

# Apply changes
source /etc/profile

# Check if Golang is installed
if command -v go &> /dev/null
then
    # Check that the installed version of Golang is at least 1.21
    GO_VERSION=$(go version | cut -d " " -f3)
    MIN_VERSION="go1.21"

    if [[ "$GO_VERSION" < "$MIN_VERSION" ]]
    then
        echo "Detected Go version $GO_VERSION. Upgrading to Go 1.21.6..."
        wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    else
        echo "Detected Go version $GO_VERSION. No upgrade needed."
    fi
else
    echo "Go is not installed. Installing Go 1.21.6..."
    wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
fi

# SET GOROOT AND GOPATH
GOROOT_LINE="export GOROOT=/usr/local/go"
GOPATH_LINE="export GOPATH=\$HOME/go"
PATH_LINE="export PATH=\$PATH:\$GOROOT/bin:\$GOPATH/bin"

# Check if GOROOT, GOPATH and PATH are already set in /etc/profile
if ! grep -q "$GOROOT_LINE" /etc/profile; then
    echo "$GOROOT_LINE" | sudo tee -a /etc/profile
fi

if ! grep -q "$GOPATH_LINE" /etc/profile; then
    echo "$GOPATH_LINE" | sudo tee -a /etc/profile
fi

if ! grep -q "$PATH_LINE" /etc/profile; then
    echo "$PATH_LINE" | sudo tee -a /etc/profile
fi

# Apply changes
source /etc/profile

# Installation ToughRADIUS
echo "Installing ToughRADIUS..."
go clean -modcache
go install github.com/talkincode/toughradius/v8@latest

# Execute the ToughRADIUS installation command
echo "Running ToughRADIUS install command..."
toughradius -install

echo "ToughRADIUS installation completed."
