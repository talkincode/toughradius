Configuration
============================

Main configuration file
-----------------------------

.. code-block:: javascript

    {
        "api" : {
            "debug" : 1,
            "host" : "0.0.0.0",
            "port" : 1815,
            "secret" : "CRTCcMB7tfnXU8aXIyfavfuqruvXkNng"
        },
        "clients" : "!include:{CONFDIR}/clients.json",
        "logger" : "!include:{CONFDIR}/logger.json",
        "radiusd" : {
            "acct_port" : 1813,
            "auth_port" : 1812,
            "adapter" : "rest",
            "debug" : 1,
            "dictionary" : "{CONFDIR}/dictionarys/dictionary",
            "free_auth_input_limit" : 1048576,
            "free_auth_output_limit" : 1048576,
            "free_auth_limit_code" : "",
            "free_auth_domain" : "",
            "host" : "0.0.0.0",
            "max_session_timeout" : 86400,
            "pass_pwd" : 0,
            "pass_userpwd" : 0,
            "pool_size" : 2,
            "request_timeout" : 5
        },
        "adapters" : {
            "rest" : {
                "url" : "http://127.0.0.1:1815/api/v1/radtest",
                "format" : "json",
                "secret" : "",
                "radattrs" : []
            }
        },
        "system" : {
            "tz" : "CST-8"
        }
    }

Logging configuration file
--------------------------------

.. code-block:: javascript

    {
        "version" : 1,
        "disable_existing_loggers" : true,
        "formatters" : {
            "verbose" : {
                "format" : "[%(asctime)s %(name)s-%(process)d] [%(levelname)s] %(message)s",
                "datefmt" : "%Y-%m-%d %H:%M:%S"
            },
            "simple" : {
                "format" : "%(asctime)s %(levelname)s %(message)s"
            }
        },
        "handlers" : {
            "null" : {
                "level" : "DEBUG",
                "class" : "logging.NullHandler"
            },
            "debug" : {
                "level" : "DEBUG",
                "class" : "logging.StreamHandler",
                "formatter" : "verbose"
            },
            "info" : {
                "level" : "DEBUG",
                "class" : "logging.handlers.TimedRotatingFileHandler",
                "when" : "d",
                "interval" : 1,
                "backupCount" : 50,
                "delay" : true,
                "filename" : "/var/log/toughradius/info.log",
                "formatter" : "verbose"
            },
            "error" : {
                "level" : "ERROR",
                "class" : "logging.handlers.TimedRotatingFileHandler",
                "when" : "d",
                "interval" : 1,
                "backupCount" : 50,
                "delay" : true,
                "filename" : "/var/log/toughradius/error.log",
                "formatter" : "verbose"
            }
        },
        "loggers" : {
            "" : {
                "handlers" : [
                    "info",
                    "error",
                    "debug"
                ],
                "level" : "DEBUG"
            }
        }
    }


Nas Client configuration file
------------------------------------

.. code-block:: javascript

    {
        "vendors" : {
            "std" : 0,
            "alcatel" : 3041,
            "cisco" : 9,
            "h3c" : 25506,
            "huawei" : 2011,
            "juniper" : 2636,
            "microsoft" : 311,
            "mikrotik" : 14988,
            "openvpn" : 19797
        },
        "defaults" : {
            "127.0.0.1" : {
                "nasid" : "toughac",
                "secret" : "secret",
                "coaport" : 3799,
                "vendor" : "std"
            }
        }
    }
