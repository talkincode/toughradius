# coding=utf-8

import os
import time
from bsddb import db
from bsddb import dbshelve
from .base import BasicAdapter
from toughradius.common import tools
from hashlib import md5
import urllib2
import json

class ShelveError(BaseException):pass

class ShelveAdapter(BasicAdapter):

    def __init__(self, config):
        BasicAdapter.__init__(self,config)
        dbpath = config.adapters.shevel.datadir
        dbfile = os.path.join(dbpath, "radius.db")
        dbenv = db.DBEnv()
        dbenv.set_shm_key(2)
        dbenv.open(dbpath, db.DB_CREATE | db.DB_INIT_MPOOL | db.DB_INIT_LOCK | db.DB_INIT_LOG | db.DB_INIT_TXN)
        self.dbm = dbshelve.open(dbfile, db.DB_RECOVER | db.DB_RECOVER_FATAL, dbenv=dbenv)

    def auth(self,req):
        try:
            pass
        except:
            self.logger.exception("send rest request error")
            raise ShelveError("rest request error")


    def acct(self,req):
        try:
            pass
        except:
            self.logger.exception("send rest request error")
            raise ShelveError("rest request error")