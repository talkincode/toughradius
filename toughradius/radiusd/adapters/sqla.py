#!/usr/bin/env python
#coding:utf-8

from .base import BasicAdapter
from toughradius.common import tools,dbengine
from toughradius.radiusd import radutils
from sqlalchemy.ext.automap import automap_base
from sqlalchemy.orm import Session
from hashlib import md5
import urllib2
import json

Base = automap_base()

STATUS_TYPE_START   = 1
STATUS_TYPE_STOP    = 2
STATUS_TYPE_UPDATE  = 3
STATUS_TYPE_UNLOCK = 4
STATUS_TYPE_CHECK_ONLINE = 5
STATUS_TYPE_ACCT_ON  = 7
STATUS_TYPE_ACCT_OFF = 8

class SqlaError(BaseException):pass

class SqlaModel(object):

    User = None
    Nas = None
    Online = None

    def __init__(self, config):
        BasicAdapter.__init__(self,config)
        self.db_engine = dbengine.get_engine(config['database'])
        Base.prepare(engine, reflect=True)
        SqlaModel.User = Base.classes.user
        SqlaModel.Online = Base.classes.online

    def query_user(self,username):
        _User = SqlaModel.User
        db = Session(self.db_engine)
        return db.query(_User).filter(_User.username==username).first()
        
    def query_online(self,username):
        _Online = SqlaModel.Online
        db = Session(self.db_engine)
        return db.query(_Online).filter(_Online.username==username).first()

class SqlaAdapter(BasicAdapter,SqlaModel):


    def auth(self,req):
        user = self.query_user(req.get_user_name())
        if not user:
            return dict(code=1, msg='user not exists')
        if user.check_passwd = 1 and  not req.is_valid_pwd(user.password):
            return dict(code=1, msg='user password error')

        prereply = dict(code=0,msg='ok')
        prereply['input_rate'] = user.input_rate
        prereply['output_rate'] = user.output_rate
        prereply['attrs']['Session-Timeout'] = radutils.calc_session_time(user.expire_date)
        if user.ip_addr:
            prereply['attrs']['Framed-IP-Address'] = user.ip_addr
        elif user.ip_pool:
            prereply['attrs']['Framed-Pool'] = user.ip_pool
        return prereply

    def acct(self,req):
        status_type = req.get_acct_status_type()
        if status_type == STATUS_TYPE_START:
            self.acct_start(req)
        elif status_type == STATUS_TYPE_UPDATE:
            self.acct_update(req)
        elif status_type == STATUS_TYPE_STOP:
            self.acct_stop(req)
        
    def acct_start(self,req):
        user = self.query_user(req.get_user_name())


    def acct_update(self,req):
        pass


    def acct_stop(self,req):
        pass

