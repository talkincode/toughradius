#!/usr/bin/env python
# coding=utf-8

from toughradius import models
from txradius import message
from toughradius.common import logger
from toughradius.radiusd import plugins

DICTIONARY = os.path.join(os.path.dirname(toughradius.__file__), 'dictionarys/dictionary')

class WorkerBasic:

    auth_req_plugins = []
    acct_req_plugins = []
    auth_accept_plugins = []
    acct_type_desc = {
        STATUS_TYPE_START: u"start",
        STATUS_TYPE_STOP: u"stop",
        STATUS_TYPE_UPDATE: u"update"
    }    

    def load_plugins(self,load_types=[]):
        def add_plugin(pobj):
            if hasattr(pobj,'plugin_types') \
                and hasattr(pobj,'plugin_name') \
                and hasattr(pobj,'plugin_priority') \
                and hasattr(pobj,'plugin_func'):
                for ptype in load_types:
                    if ptype == 'radius_auth_req' and ptype in pobj.plugin_types \
                        and pobj not in self.auth_req_plugins:
                        self.auth_req_plugins.append(pobj)
                        logger.info('load auth_req_plugin -> {} {}'.format(pobj.plugin_priority,pobj.plugin_name))
                    if ptype == 'radius_accept' and ptype in pobj.plugin_types \
                        and pobj not in self.auth_accept_plugins:
                        self.auth_accept_plugins.append(pobj)
                        logger.info('load auth_accept_plugin -> {} {}'.format(pobj.plugin_priority,pobj.plugin_name))
                    if ptype == 'radius_acct_req' and ptype in pobj.plugin_types \
                        and pobj not in self.acct_req_plugins:
                        self.acct_req_plugins.append(pobj)
                        logger.info('load acct_req_plugin -> {} {}'.format(pobj.plugin_priority,pobj.plugin_name))
                    

        default_plugins_dir = os.path.dirname(plugins.__file__)
        logger.info('start load radius plugins {} from {}'.format(load_types,default_plugins_dir))
        modules = (os.path.splitext(m)[0] for m in os.listdir(default_plugins_dir))
        for pgname in modules:
            try:
                pg_prefix = 'toughradius.radius.plugins'
                add_plugin(importlib.import_module('{0}.{1}'.format(pg_prefix,pgname)))
            except Exception as err:
                logger.exception(err,trace="radius",tag="radius_plugins_load_error")

        self.auth_req_plugins = sorted(self.auth_req_plugins,key=lambda i:i.plugin_priority)
        self.acct_req_plugins = sorted(self.acct_req_plugins,key=lambda i:i.plugin_priority)
        self.auth_accept_plugins = sorted(self.auth_accept_plugins,key=lambda i:i.plugin_priority)  

    def get_param_value(self, name, defval=None):
        def fetch_result():
            table = models.TrParam.__table__
            with self.db_engine.begin() as conn:
                return conn.execute(
                    table.select().with_only_columns([table.c.param_value]).where(
                        table.c.param_name==name)).scalar() or defval
        try:
            return self.mcache.aget(param_cache_key(name),fetch_result, expire=600)
        except Exception as err:
            logger.exception(err,trace="radius")
            return defval

    def is_trace_on(self):
        return int(self.get_param_value('radius_user_trace',0))

    def user_exists(self,username):
        def fetch_result():
            table = models.TrAccount.__table__
            with self.db_engine.begin() as conn:
                val = conn.execute(table.select().where(
                    table.c.account_number==username)).first()
                return val and Storage(val.items()) or None
        return self.mcache.aget(account_cache_key(username),fetch_result, expire=600) is not None

    def log_trace(self,host,port,req,reply=None):
        """ Tracking logging, need to set the global config
        """
        if not self.is_trace_on():
            return
        if not self.user_exists(req.get_user_name()):
            return

        try:
            if reply is None:
                if req.code == packet.AccessRequest:
                    logger.info(u"User:%s authorize request, Usermac=%s"%(
                        req.get_user_name(),
                        req.get_mac_addr()),
                        trace="radius", username=req.get_user_name()
                    )
                elif req.code == packet.AccountingRequest:
                    logger.info(u"User:%s accounting (%s) request, Usermac=%s"%(
                        req.get_user_name(),
                        self.acct_type_desc.get(req.get_acct_status_type(),''),
                        req.get_mac_addr()),
                        trace="radius",username=req.get_user_name()
                    )

                msg = message.format_packet_str(req)
                logger.info(u"[RADIUSD] Receive Radius request from nas (%s:%s)  %s"%(
                    host,port,utils.safeunicode(msg)),
                    trace="radius",username=req.get_user_name()
                )
            else:
                if reply.code == packet.AccessReject:
                    logger.info(u'User Authentication is rejected: %s' % reply['Reply-Message'][0], tag="radius_auth_reject", 
                        eslog=True, trace="radius",username=req.get_user_name())
                elif reply.code == packet.AccessAccept:
                    logger.info(reply['Reply-Message'][0],trace="radius",username=req.get_user_name())
                elif reply.code == packet.AccountingResponse:
                    logger.info(u"User Success of account",trace="radius",username=req.get_user_name())

                msg = message.format_packet_str(reply)
                logger.info(u"[RADIUSD] Send Radius response to nas (%s:%s)  %s"%(
                    host,port,utils.safeunicode(msg)),
                    trace="radius",username=req.get_user_name()
                )
        except Exception as err:
            logger.exception(err)


    def find_nas(self,ip_addr):
        """ Query nas device based on IP address
        """
        def fetch_result():
            table = models.TrBas.__table__
            with self.db_engine.begin() as conn:
                return conn.execute(table.select().where(table.c.ip_addr==ip_addr)).first()
        return self.mcache.aget(bas_cache_key(ip_addr),fetch_result, expire=600)


    def get_account_bind_nas(self,account_number):
        def fetch_result():
            with self.db_engine.begin() as conn:
                tbas = models.TrBas.__table__
                tcus = models.TrCustomer.__table__
                tuser = models.TrAccount.__table__
                tbn = models.TrBasNode.__table__
                with self.db_engine.begin() as conn:
                    stmt = tbas.select().with_only_columns([tbas.c.ip_addr]).where(
                            tcus.c.customer_id==tuser.c.customer_id).where(
                            tcus.c.node_id==tbn.c.node_id).where(
                            tbn.c.bas_id==tbas.c.id).where(
                            tuser.c.account_number==account_number)

                    vals = conn.execute(stmt)
                    ipaddrs = [ addr.ip_addr for addr in vals]
                return ipaddrs
        return self.mcache.aget(account_bind_bas_key(account_number),fetch_result, expire=600)


    def do_auth_stat(self,code):
        try:
            stat_msg = {'statattrs':[],'raddata':{}}
            if code == packet.AccessRequest:
                stat_msg['statattrs'].append('auth_req')
            elif code == packet.AccessAccept:
                stat_msg['statattrs'].append('auth_accept')
            elif  code == packet.AccessReject:
                stat_msg['statattrs'].append('auth_reject')
            else:
                stat_msg['statattrs'] = ['auth_drop']
            self.stat_pusher.push(msgpack.packb(stat_msg))
        except:
            pass

    def do_acct_stat(self,code, status_type=0, req=None):
        try:
            stat_msg = {'statattrs':['acct_drop'],'raddata':{}}
            if code  in (4,5):
                stat_msg['statattrs'] = []
                if code == packet.AccountingRequest:
                    stat_msg['statattrs'].append('acct_req')
                elif code == packet.AccountingResponse:
                    stat_msg['statattrs'].append('acct_resp')

                if status_type == 1:
                    stat_msg['statattrs'].append('acct_start')
                elif status_type == 2:
                    stat_msg['statattrs'].append('acct_stop')        
                elif status_type == 3:
                    stat_msg['statattrs'].append('acct_update')   
                    stat_msg['raddata']['input_total'] = req.get_input_total()     
                    stat_msg['raddata']['output_total'] = req.get_output_total()     
                elif status_type == 7:
                    stat_msg['statattrs'].append('acct_on')        
                elif status_type == 8:
                    stat_msg['statattrs'].append('acct_off')

            self.stat_pusher.push(msgpack.packb(stat_msg))
        except:
            pass
