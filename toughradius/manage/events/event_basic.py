#!/usr/bin/env python
#coding:utf-8

from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.manage import models
from toughlib.dbutils import make_db
from toughradius.manage.settings import param_cache_key

class BasicEvent:

    def __init__(self, dbengine=None, mcache=None, aes=None,**kwargs):
        self.dbengine = dbengine
        self.mcache = mcache
        self.db = scoped_session(sessionmaker(bind=self.dbengine, autocommit=False, autoflush=False))
        self.aes = aes
        
    def get_param_value(self, name):
        def fetch_result():
            table = models.TrParam.__table__
            with self.dbengine.begin() as conn:
                return conn.execute(table.select().with_only_columns([table.c.param_value]).where(
                        table.c.param_name==name)).scalar()
        return self.mcache.aget(param_cache_key(name),fetch_result, expire=300)


    def get_customer_info(self, account_number):
        with make_db(self.db) as db:
            return db.query(
                models.TrCustomer.mobile,
                models.TrCustomer.realname,
                models.TrProduct.product_name,
                models.TrAccount.account_number,
                models.TrAccount.install_address,
                models.TrAccount.expire_date,
                models.TrAccount.password
            ).filter(
                models.TrCustomer.customer_id == models.TrAccount.customer_id,
                models.TrAccount.product_id == models.TrProduct.id,
                models.TrAccount.account_number == account_number
            ).first()



__call__ = lambda **kwargs: None

