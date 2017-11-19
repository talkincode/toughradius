# TOUGHRADIUS

[![Build Status](https://travis-ci.org/talkincode/ToughRADIUS.svg?branch=master)](https://travis-ci.org/talkincode/ToughRADIUS)

ToughRADIUS is a Radius server software developed based on Python, which implements the standard Radius protocol and supports the extension of Radius protocol.

ToughRADIUS can be understood as a Radius middleware, and it does not implement all of the business functions. But it's easy to Easier to extended development.

ToughRADIUS is similar to freeRADIUS, But it's simpler to use, Easier to extended development.

# Quick start

## Install

First of all, you need a Python runtime environment and install the pip package manage tool

    > pip install toughradius

## Startup

Start authentication, accounting listen service

    >> radiusd

debug mode

    >> radiusd -x
    
for help 

    >> radiusd --help