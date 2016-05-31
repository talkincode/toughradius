#!/usr/bin/env python
#coding:utf-8
import sys, struct
from toughlib import utils,httpclient
from toughlib import dispatch,logger
from toughradius.manage import models
from toughlib.dbutils import make_db
from toughradius.manage.tasks.task_base import TaseBasic
from twisted.internet import reactor,defer
from twisted.names import client, dns
from toughradius.manage import taskd

class DdnsUpdateTask(TaseBasic):

    __name__ = 'ddns-update'

    def first_delay(self):
        return 5

    def get_notify_interval(self):
        return 60

    @defer.inlineCallbacks
    def process(self, *args, **kwargs):
        self.logtimes()
        with make_db(self.db) as db:
            try:
                nas_list = db.query(models.TrBas)
                for nas in nas_list:
                    if not nas.dns_name:
                        continue
                    results, _, _ = yield client.lookupAddress(nas.dns_name)
                    if not results:
                        logger.info("domain {0} resolver empty".format(nas.dns_name))

                    if results[0].type == dns.A:
                        ipaddr = ".".join(str(i) for i in struct.unpack("BBBB", results[0].payload.address))
                        if ipaddr:
                            nas.ip_addr = ipaddr
                            db.commit()
                            logger.info("domain {0} resolver {1}  success".format(nas.dns_name,ipaddr))
                    else:
                        logger.info("domain {0} no ip address,{1}".format(nas.dns_name, repr(results)))

            except Exception as err:
                logger.exception(err)
        defer.returnValue(self.get_notify_interval())


taskd.TaskDaemon.__taskclss__.append(DdnsUpdateTask)




