#!/usr/bin/env python
# coding:utf-8

import os
import ConfigParser


class ConfigDict(dict):
    def __getattr__(self, key):
        try:
            return self[key]
        except KeyError, k:
            raise AttributeError, k

    def __setattr__(self, key, value):
        self[key] = value

    def __delattr__(self, key):
        try:
            del self[key]
        except KeyError, k:
            raise AttributeError, k

    def __repr__(self):
        return '<ConfigDict ' + dict.__repr__(self) + '>'


class Config():
    """ Config Object """

    def __init__(self, conf_file=None, **kwargs):

        cfgs = [conf_file, '/etc/radiusd.conf']
        self.config = ConfigParser.ConfigParser()
        flag = False
        for c in cfgs:
            if c and os.path.exists(c):
                self.config.read(c)
                self.filename = c
                flag = True
                break
        if not flag:
            raise Exception("no config")

        self.defaults = ConfigDict(**{k: v for k, v in self.config.items("DEFAULT")})
        self.memcached = ConfigDict(**{k: v for k, v in self.config.items("memcached") if k not in self.defaults})
        self.admin = ConfigDict(**{k: v for k, v in self.config.items("admin") if k not in self.defaults})
        self.database = ConfigDict(**{k: v for k, v in self.config.items("database") if k not in self.defaults})

        self.update_boolean()
        self.setup_env()

    def update_boolean(self):
        self.defaults.debug = self.defaults.debug in ("1", "true")
        self.database.echo = self.database.echo in ("1", "true")

    def setup_env(self):

        _syslog_enable = os.environ.get("SYSLOG_ENABLE")
        _syslog_server = os.environ.get("SYSLOG_SERVER")
        _syslog_port = os.environ.get("SYSLOG_PORT")
        _syslog_level = os.environ.get("SYSLOG_LEVEL")
        _timezone = os.environ.get("TIMEZONE")
        _db_type = os.environ.get("DB_TYPE")
        _db_url = os.environ.get("DB_URL")
        _memcached_hosts = os.environ.get("MEMCACHED_HOSTS")
        _zauth_port = os.environ.get("ZAUTH_PORT")
        _zacct_port = os.environ.get("ZACCT_PORT")

        if _syslog_enable:
            self.defaults.syslog_enable = _syslog_enable
        if _syslog_server:
            self.defaults.syslog_server = _syslog_server
        if _syslog_port:
            self.defaults.syslog_port = _syslog_port
        if _syslog_level:
            self.defaults.syslog_level = _syslog_level
        if _timezone:
            self.defaults.tz = _timezone
        if _db_type:
            self.database.dbtype = _db_type
        if _db_url:
            self.database.dburl = _db_url
        if _memcached_hosts:
            self.memcached.hosts = _memcached_hosts             
        if _zauth_port:
            self.admin._zauth_port = _zauth_port                  
        if _zacct_port:
            self.admin._zacct_port = _zacct_port        


    def update(self):
        """ update config file"""
        for k, v in self.defaults.iteritems():
            self.config.set("DEFAULT", k, v)
                
        for k, v in self.memcached.iteritems():
            if k not in self.defaults:
                self.config.set("memcached", k, v)

        for k, v in self.admin.iteritems():
            if k not in self.defaults:
                self.config.set("admin", k, v)

        for k, v in self.database.iteritems():
            if k not in self.defaults:
                self.config.set("database", k, v)

        with open(self.filename, 'w') as cfs:
            self.config.write(cfs)

        self.update_boolean()


def find_config(conf_file=None):
    return Config(conf_file)

if __name__ == "__main__":
    pass





