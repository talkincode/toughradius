# TOUGHRADIUS

[![Build Status](https://travis-ci.org/talkincode/ToughRADIUS.svg?branch=master)](https://travis-ci.org/talkincode/ToughRADIUS)

TOUGHRADIUS is a Radius server software developed based on Python, which implements the standard Radius protocol and supports the extension of Radius protocol.

TOUGHRADIUS can be understood as a Radius middleware, and it does not implement all of the business functions. It needs access to the back-end business system.

TOUGHRADIUS provides complete back end interface support, such as supporting HTTP protocol access capabilities, and the back-end business system must provide additional HTTP API interfaces,


# Quick start

## Install

First of all, you need a Python runtime environment and install the pip package manage tool

    > pip install toughradius

## Configuration

The default configuration directory is /etc/toughradius

- Main configuration file: /etc/toughradius/radiusd.json
- Logging configuration file: /etc/toughradius/radiusd.json
- Nas Client configuration file: /etc/toughradius/clients.json
- Radius protocol dictionary file directory: /etc/toughradius/dictionarys

## Startup

Start authentication, accounting, and API Server on one process

    > gtrcli radiusd

Launching authentication services only

    > gtrcli auth

Launching accounting services only

    > gtrcli acct

Launching API Server  only

    > gtrcli apiserv
    