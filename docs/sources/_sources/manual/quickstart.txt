Quick start
=======================

1. Install
------------

First of all, you need a Python runtime environment and install the pip package manage tool


::

    pip install toughradius

2. Configuration
---------------------

The default configuration directory is /etc/toughradius

- Main configuration file: /etc/toughradius/radiusd.json
- Logging configuration file: /etc/toughradius/radiusd.json
- Nas Client configuration file: /etc/toughradius/clients.json
- Radius protocol dictionary file directory: /etc/toughradius/dictionarys

3. Startup
----------------

Start authentication, accounting, and API Server on one process

::

    gtrcli radiusd

Launching authentication services only

::

    gtrcli auth

Launching accounting services only

::

    gtrcli acct

Launching API Server  only

::

    gtrcli apiserv