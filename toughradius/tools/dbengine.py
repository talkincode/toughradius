#!/usr/bin/env python
#coding:utf-8
import functools
from sqlalchemy import create_engine
ISOLATION_LEVEL = {
    1 : 'READ COMMITTED',
    2 : 'READ UNCOMMITTED',
    3 : 'REPEATABLE READ',
    4 : 'SERIALIZABLE'
}


class DBEngine(object):

    def __init__(self,config):
        self.config = config
        self.conf_str = functools.partial(self.config.get,"database")
        self.conf_int = functools.partial(self.config.getint,"database")
        self.conf_bool = functools.partial(self.config.getboolean,"database")
        
    def __call__(self):
        return self.get_engine()

    def get_engine(self):
        if self.conf_str('dbtype') == 'mysql':
            return create_engine(
                self.conf_str("dburl"),
                echo=self.conf_bool('echo'),
                pool_size = self.conf_int('pool_size'),
                pool_recycle=self.conf_int('pool_recycle')
            )
        elif self.conf_str('dbtype') == 'postgresql':
            return create_engine(
                self.conf_str("dburl"),
                echo=self.conf_bool('echo'),
                pool_size = self.conf_int('pool_size'),
                isolation_level = ISOLATION_LEVEL.get(self.conf_int('isolation_level'),1),
                pool_recycle=self.conf_int('pool_recycle')
            )
        elif self.conf_str('dbtype') == 'sqlite':
            def my_con_func():
                import sqlite3.dbapi2 as sqlite
                con = sqlite.connect(self.conf_str("dburl").replace('sqlite:///',''))
                con.text_factory=str
                # con.execute("PRAGMA synchronous=OFF;")
                # con.isolation_level = 'IMMEDIATE'
                return con
            return create_engine(
                "sqlite+pysqlite:///",
                creator=my_con_func,
                echo=self.conf_bool('echo')
            )
        else:
            return create_engine(
                self.conf_str("dburl"),
                echo=self.conf_bool('echo'),
                pool_size = self.conf_int('pool_size')
            )

def get_engine(config):
    return DBEngine(config)()