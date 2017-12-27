# coding:utf-8

import os

ENVIRONMENT_VARIABLE = "TOUGHRADIUS_SETTINGS_MODULE"
BASICDIR = os.path.abspath(os.path.dirname(__file__))

'''
define nas access devices vendor ids
'''

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

'''
- debug: enable debug mode
- host: radius server listen host
- auth_port: radius auth listen port
- acct_port: radius acct listen port
- adapter: radius handle adapter module
- dictionary: include an additional  Radius protocol dictionary file directory path
- debug: debug model setting
- pool_size: radius server worker pool size
'''

RADIUSD = {
    "host": "0.0.0.0",
    "auth_port": 1812,
    "acct_port": 1813,
    # "adapter": "toughradius.radiusd.adapters.rest",
    "adapter": "toughradius.radiusd.adapters.rest",
    "debug": 1,
    "dictionary": os.path.join(BASICDIR,'dictionarys/dictionary'),
    "pool_size": 128
}

'''
default rest adapter module config
- authurl: backend server authentication api url
- accturl: backend server accounting api url
- secret: http message sign secret
- radattrs: Radius attrs send to  backend server
'''

ADAPTERS = {
    "rest" : {
        "nasurl" : "http://127.0.0.1:1879/api/nas/get",
        "authurl" : "http://127.0.0.1:1879/api/radius/auth",
        "accturl" : "http://127.0.0.1:1879/api/radius/acct",
        "appid" : "IhFkbPG4lcNoTYXF",
        "secret" : "ASTRGZkFD4Ri94I3JNvKYmIL0GadxfBK",
    }
}

'''
radius ext modules 
'''

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

'''
- radius server logging config
'''

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