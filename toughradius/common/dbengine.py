#!/usr/bin/env python
#coding:utf-8
import os
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
        self.dbtype = os.environ.get("DB_TYPE", self.config.database.dbtype)
        self.dburl = os.environ.get("DB_URL", self.config.database.dburl)

    def __call__(self):
        return self.get_engine()

    def get_engine(self):
        if self.dbtype == 'mysql':
            return create_engine(
                self.dburl,
                echo=bool(self.config.database.echo),
                pool_size = int(self.config.database.pool_size),
                pool_recycle=int(self.config.database.pool_recycle)
            )
        elif self.dbtype == 'postgresql':
            return create_engine(
                self.dburl,
                echo=bool(self.config.database.echo),
                pool_size = int(self.config.database.pool_size),
                isolation_level = int(ISOLATION_LEVEL.get(self.config.database.isolation_level, 1)),
                pool_recycle=int(self.config.database.pool_recycle)
            )
        elif self.dbtype == 'sqlite':
            def my_con_func():
                import sqlite3.dbapi2 as sqlite
                con = sqlite.connect(self.dburl.replace('sqlite:///',''))
                con.text_factory=str
                # con.execute("PRAGMA synchronous=OFF;")
                # con.isolation_level = 'IMMEDIATE'
                return con
            return create_engine(
                "sqlite+pysqlite:///",
                creator=my_con_func,
                echo=bool(self.config.database.echo)
            )
        else:
            return create_engine(
                self.dburl,
                echo=bool(self.config.database.echo),
                pool_size = int(self.config.database.pool_size)
            )

def get_engine(config):
    return DBEngine(config)()