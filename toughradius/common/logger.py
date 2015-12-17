#!/usr/bin/env python
# coding=utf-8
import sys
import socket
import logging
import logging.handlers


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


class Logger:

    def __init__(self,config):
        self.formatter = logging.Formatter(
            u'%(asctime)s {0} %(name)s %(levelname)-8s %(module)s %(message)s'.format(config.defaults.syslog_shost),
            '%b %d %H:%M:%S', )
        self.syslog_enable = config.defaults.get("syslog_enable") in ('1', 'true', 'on')
        self.syslog_server = config.defaults.get('syslog_server')
        self.syslog_port = int(config.defaults.get('syslog_port', 514))
        if self.syslog_server:
            self.syslog_address = (self.syslog_server,self.syslog_port)
        self.level = string_to_level(config.defaults.get('syslog_level', 'INFO'))
        if config.defaults.debug:
            self.level = string_to_level("DEBUG")

        self.syslogger = logging.getLogger('toughadmin')
        self.syslogger.setLevel(self.level)

        if self.syslog_enable and self.syslog_server:
            handler = logging.handlers.SysLogHandler(address=(self.syslog_server, self.syslog_port))
            handler.setFormatter(self.formatter)
            self.syslogger.addHandler(handler)

        if config.defaults.debug:
            stream_handler = logging.StreamHandler(sys.stderr)
            stream_handler.setFormatter(self.formatter)
            self.syslogger.addHandler(stream_handler)

        self.info = self.syslogger.info
        self.debug = self.syslogger.debug
        self.warning = self.syslogger.warning
        self.error = self.syslogger.error
        self.critical = self.syslogger.critical
        self.log = self.syslogger.log
        self.msg = self.syslogger.info
        self.err = self.syslogger.error

