#!/usr/bin/env python
#coding:utf-8
from toughlib.dbutils import make_db
from toughradius.manage import models

class TaseBasic:

    def __init__(self, config=None, db=None, **kwargs):
        self.config = config
        self.db = db

    def process(self,*args, **kwargs):
        pass

    def get_param_value(self, name, defval=None):
        with make_db(self.db) as conn:
            val = self.db.query(models.TrParam.param_value).filter_by(param_name = name).scalar()
            return val or defval