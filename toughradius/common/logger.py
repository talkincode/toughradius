#!/usr/bin/env python
# coding=utf-8
import sys
import os
import socket
import traceback
import logging
import logging.handlers
from toughradius.common import dispatch
from toughradius.common.utils import safeunicode
from twisted.python import log as txlog
import functools

EVENT_TRACE = 'syslog_trace'
EVENT_INFO = 'syslog_info'
EVENT_DEBUG = 'syslog_debug'
EVENT_ERROR = 'syslog_error'
EVENT_EXCEPTION = 'syslog_exception'
EVENT_SETUP = 'syslog_setup'


def string_to_level(log_level):
    if log_level == "CRITICAL":
        return logging.CRITICAL
    if log_level == "ERROR":
        return logging.ERROR
    if log_level == "WARNING":
        return logging.WARNING
    if log_level == "INFO":
        return logging.INFO
    if log_level == "DEBUG":
        return logging.DEBUG
    return logging.NOTSET

class SimpleLogger:

    def __init__(self,config, name="toughstruct"):
        self.name = name
        self.setup(config)

    def setup(self, config):
        self.level = string_to_level(config.syslog.level)
        if config.system.debug:
            self.level = string_to_level("DEBUG")

        self.log = logging.getLogger(self.name)
        self.log.setLevel(self.level)

        handler = logging.StreamHandler(sys.stdout)
        formatter = logging.Formatter(u'%(message)s')
        handler.setFormatter(formatter)
        self.log.addHandler(handler)

        self.info = self.log.info
        self.debug = self.log.debug
        self.warning = self.log.warning
        self.error = self.log.error
        self.critical = self.log.critical
        self.msg = self.log.info
        self.err = self.log.error

    def event_syslog_setup(self,config):
        self.setup(config)

    def event_syslog_info(self, msg):
        self.info(msg)

    def event_syslog_debug(self, msg):
        self.debug(msg)

    def event_syslog_error(self, msg):
        self.error(msg)

    def event_syslog_exception(self, err):
        self.log.exception(err)        

    def emit(self, eventDict):
        text = txlog.textFromEventDict(eventDict)
        if text is None:
            return
        if eventDict['isError'] and 'failure' in eventDict:
            self.error(text)
        else:
            self.info(text)


class Logger:

    def __init__(self,config, name="toughstruct"):
        self.name = name
        self.setup(config)

    def setup(self, config):
        self.syslog_enable = config.syslog.enable
        self.syslog_server = config.syslog.server
        self.syslog_port = config.syslog.port
        self.syslog_level = config.syslog.level
        self.syslog_shost = config.syslog.shost
        self.formatter = logging.Formatter(
            u'%(asctime)s {0} %(name)s %(levelname)-8s %(message)s'.format(self.syslog_shost),'%b %d %H:%M:%S', )
        self.level = string_to_level(self.syslog_level)
        if config.system.debug:
            self.level = string_to_level("DEBUG")

        self.syslogger = logging.getLogger(self.name)
        self.syslogger.setLevel(self.level)

        if self.syslog_enable and self.syslog_server:
            handler = logging.handlers.SysLogHandler(address=(self.syslog_server, self.syslog_port))
            handler.setFormatter(self.formatter)
            self.syslogger.addHandler(handler)
        else:
            handler = logging.StreamHandler(sys.stderr)
            formatter = logging.Formatter(u'\x1b[32;40m[%(asctime)s %(name)s]\x1b[0m %(message)s','%b %d %H:%M:%S',)
            handler.setFormatter(formatter)
            self.syslogger.addHandler(handler)

        self.info = self.syslogger.info
        self.debug = self.syslogger.debug
        self.warning = self.syslogger.warning
        self.error = self.syslogger.error
        self.critical = self.syslogger.critical
        self.log = self.syslogger.log
        self.msg = self.syslogger.info
        self.err = self.syslogger.error

    def event_syslog_setup(self,config):
        self.setup(config)

    def event_syslog_info(self, msg, **kwargs):
        self.info(msg)

    def event_syslog_debug(self, msg, **kwargs):
        self.debug(msg)

    def event_syslog_error(self, msg, **kwargs):
        self.error(msg)

    def event_syslog_exception(self, err, **kwargs):
        self.syslogger.exception(err)

    def emit(self, eventDict):
        text = txlog.textFromEventDict(eventDict)
        if text is None:
            return
        if not isinstance(text, (unicode,str,dict,list)):
            text = text
        else:
            text = safeunicode(text)

        if eventDict['isError'] and 'failure' in eventDict:
            self.exception(text)
        else:
            self.info(text)


setup = functools.partial(dispatch.pub, EVENT_SETUP) 


def info(message,trace="info",**kwargs):
    if not isinstance(message, unicode):
        message = safeunicode(message)
    if EVENT_INFO in dispatch.dispatch.callbacks:
        dispatch.pub(EVENT_INFO,message,**kwargs)
        if EVENT_TRACE in dispatch.dispatch.callbacks:
            dispatch.pub(EVENT_TRACE,trace,message,**kwargs)


def debug(message,**kwargs):
    if not isinstance(message, unicode):
        message = safeunicode(message)
    if EVENT_DEBUG in dispatch.dispatch.callbacks:
        dispatch.pub(EVENT_DEBUG,message,**kwargs)
        if EVENT_TRACE in dispatch.dispatch.callbacks:
            dispatch.pub(EVENT_TRACE,"debug",message,**kwargs)

def error(message,**kwargs):
    if not isinstance(message, unicode):
        message = safeunicode(message)
    if EVENT_ERROR in dispatch.dispatch.callbacks:
        dispatch.pub(EVENT_ERROR,message,**kwargs)
        if EVENT_TRACE in dispatch.dispatch.callbacks:
            dispatch.pub(EVENT_TRACE,"error",message,**kwargs)



def exception(err,**kwargs):
    if EVENT_EXCEPTION in dispatch.dispatch.callbacks:
        dispatch.pub(EVENT_EXCEPTION,err,**kwargs)
        if EVENT_TRACE in dispatch.dispatch.callbacks:
            dispatch.pub(EVENT_TRACE,"exception",repr(err),**kwargs)

def trace_exception(etype, value, tb):
    errmsg = "".join(traceback.format_exception(etype, value, tb))
    error(errmsg,trace="exception")

sys.excepthook = trace_exception
