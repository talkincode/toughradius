#!/bin/bash

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
        rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    else
        echo "Detected Go version $GO_VERSION. No upgrade needed."
    fi
else
    echo "Go is not installed. Installing Go 1.21.6..."
    wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
fi

# SET GOROOT AND GOPATH
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Installation ToughRADIUS
echo "Installing ToughRADIUS..."
go install github.com/talkincode/toughradius/v8@latest

# Execute the ToughRADIUS installation command
echo "Running ToughRADIUS install command..."
toughradius -install

echo "ToughRADIUS installation completed."
