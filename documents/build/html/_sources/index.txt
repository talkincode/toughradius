.. toughradius documentation master file, created by
   sphinx-quickstart on Sat Aug 19 20:44:52 2017.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

TOUGHRADIUS User Manual
=======================================

TOUGHRADIUS is a Radius server software developed based on Python, which implements the standard Radius protocol and supports the extension of Radius protocol.

TOUGHRADIUS can be understood as a Radius middleware, and it does not implement all of the business functions. It needs access to the back-end business system.

TOUGHRADIUS provides complete back end interface support, such as supporting HTTP protocol access capabilities, and the back-end business system must provide additional HTTP API interfaces,


About this manual
----------------------

- The manual applies to TOUGHRADIUS V5.0 and later.
- For questions and omissions in the document, you can send us feedback via email (jamiesun.net@gmail.com).
- Https://github.com/talkincode/ToughRADIUS/issues feedback is  recommended.



User guide:
---------------

.. toctree::
   :maxdepth: 2

   manual/quickstart
   manual/configuration
   manual/dirstruct
   manual/deploy


Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`

