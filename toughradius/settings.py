# coding:utf-8

import os

ENVIRONMENT_VARIABLE = "TOUGHRADIUS_SETTINGS_MODULE"
BASICDIR = os.path.abspath(os.path.dirname(__file__))

VENDORS = {
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

RADIUSD = {
    "host": "0.0.0.0",
    "auth_port": 1812,
    "acct_port": 1813,
    # "adapter": "toughradius.radiusd.adapters.rest",
    "adapter": "toughradius.radiusd.adapters.free",
    "debug": 1,
    "dictionary": os.path.join(BASICDIR,'dictionarys/dictionary'),
    "pool_size": 128
}

ADAPTERS = {
    "rest" : {
        "authurl" : "http://127.0.0.1:1815/api/v1/radtest",
        "accturl" : "http://127.0.0.1:1815/api/v1/radtest",
        "secret" : "",
        "radattrs" : []
    }
}

MODULES = {
    "auth_pre" : [
        "toughradius.radiusd.modules.request_logger",
        "toughradius.radiusd.modules.request_mac_parse",
        "toughradius.radiusd.modules.request_vlan_parse"
    ],

    "acct_pre" : [
        "toughradius.radiusd.modules.request_logger",
        "toughradius.radiusd.modules.request_mac_parse",
        "toughradius.radiusd.modules.request_vlan_parse"
    ],

    "auth_post" : [
        "toughradius.radiusd.modules.response_logger",
        "toughradius.radiusd.modules.accept_rate_process"
    ],

    "acct_post" : [
        "toughradius.radiusd.modules.response_logger",
    ],
}

LOGGER = {
    "version" : 1,
    "disable_existing_loggers" : True,
    "formatters" : {
        "verbose" : {
            "format" : "[%(asctime)s %(name)s-%(process)d] [%(levelname)s] %(message)s",
            "datefmt" : "%Y-%m-%d %H:%M:%S"
        },
        "simple" : {
            "format" : "%(asctime)s %(levelname)s %(message)s"
        },
        "json": {
            '()': 'toughradius.common.json_log_formater.JSONFormatter'
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
        },
        "accounting": {
            "level": "INFO",
            "class": "logging.handlers.TimedRotatingFileHandler",
            "when": "d",
            "interval": 1,
            "backupCount": 30,
            "delay": True,
            "filename": "/var/log/toughradius/accounting.log",
            "formatter": "json"
        },
        "ticket": {
            "level": "INFO",
            "class": "logging.handlers.TimedRotatingFileHandler",
            "when": "d",
            "interval": 1,
            "backupCount": 30,
            "delay": True,
            "filename": "/var/log/toughradius/ticket.log",
            "formatter": "json"
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
        },
        "accounting" : {
            'handlers': ['accounting'],
            'level': 'INFO',
        },
        "ticket" : {
            'handlers': ['ticket'],
            'level': 'INFO',
        },
    }
}