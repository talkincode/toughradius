Development directory structure
======================================

Understanding the development directory structure helps to quickly understand the system and facilitate further expansion of development.

.. code-block:: bash

    ├── docs   # User manual documentation
    ├── etc # Configuration directory
    │   ├── clients.json
    │   ├── dictionarys  # Radius protocol dictionary file directory
    │   ├── logger.json
    │   ├── radiusd.conf
    │   └── radiusd.json
    └── toughradius  # Main source code package
    ├── setup.py  # Package script
    ├── debug.py # Local development, debugging, running scripts
    ├── requirements.txt  # A list of Python modules used in the development environment
    ├── Changelog.md  # Change logs
    ├── LICENSE  # Software license agreement file
    ├── Makefile # Software system description
    ├── README.md


