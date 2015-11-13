#!/usr/bin/env python
# coding=utf-8
import sys
import socket
import logging
import logging.handlers

FORMATTER = logging.Formatter(u'%(asctime)s {0} %(name)s %(levelname)s %(module)s -> %(funcName)s (%(lineno)d) %(message)s'.format(socket.gethostname()), '%b %d %H:%M:%S', )

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


class SysLogger:

    def __init__(self,config):
        self.syslog_enable = config.getboolean("DEFAULT", "syslog_enable")
        self.syslog_server = config.get("DEFAULT", 'syslog_server')
        self.syslog_port = config.getint("DEFAULT", 'syslog_port') or 514
        if self.syslog_server:
            self.syslog_address = (self.syslog_server,self.syslog_port)
        self.level = logging.INFO
        if config.getboolean("DEFAULT", "debug"):
            self.level = logging.DEBUG
        self.syslogger = logging.getLogger('toughradius')
        self.syslogger.setLevel(self.level)

        if self.syslog_enable and self.syslog_server:
            try:
                handler = logging.handlers.SysLogHandler(address=(self.syslog_server, self.syslog_port))
                handler.setFormatter(FORMATTER)
                self.syslogger.addHandler(handler)
                print 'enable syslog logging'
            except:
                pass

        print 'enable basic logging'
        stream_handler = logging.StreamHandler(sys.stderr)
        formatter = logging.Formatter(u'%(name)-7s %(asctime)s %(levelname)-8s %(message)s',
                                      '%a, %d %b %Y %H:%M:%S', )
        stream_handler.setFormatter(formatter)
        self.syslogger.addHandler(stream_handler)

        self.info = self.syslogger.info
        self.debug = self.syslogger.debug
        self.warn = self.syslogger.warning
        self.error = self.syslogger.error
        self.crit = self.syslogger.critical
        self.exception = self.syslogger.exception


