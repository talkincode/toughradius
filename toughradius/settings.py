# coding:utf-8

import os

__basedir__ = os.path.abspath(os.path.dirname(__file__))

vendors = {
    "std" : 0,
    "alcatel" : 3041,
    "cisco" : 9,
    "h3c" : 25506,
    "huawei" : 2011,
    "juniper" : 2636,
    "microsoft" : 311,
    "mikrotik" : 14988,
    "openvpn" : 19797
}

radiusd = {
    "host": "0.0.0.0",
    "auth_port": 1812,
    "acct_port": 1813,
    "adapter": "rest",
    "debug": 0,
    "dictionary": os.path.join(__basedir__,'dictionarys/dictionary'),
    "pool_size": 128
}

adapters = {
    "rest" : {
        "authurl" : "http://127.0.0.1:1815/api/v1/radtest",
        "accturl" : "http://127.0.0.1:1815/api/v1/radtest",
        "secret" : "",
        "radattrs" : []
    }
}

logger = {
    "version" : 1,
    "disable_existing_loggers" : True,
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
            "backupCount" : 30,
            "delay" : True,
            "filename" : "/var/log/toughradius/info.log",
            "formatter" : "verbose"
        },
        "error" : {
            "level" : "ERROR",
            "class" : "logging.handlers.TimedRotatingFileHandler",
            "when" : "d",
            "interval" : 1,
            "backupCount" : 30,
            "delay" : True,
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