#!/usr/bin/env python
# coding=utf-8

def run_auth(config):
    from twisted.internet import reactor
    from toughradius.radiusdd.server.master import RADIUSMaster
    auth_protocol = RADIUSMaster(config, service='auth')
    reactor.listenUDP(int(config.radiusd.auth_port), auth_protocol, interface=config.radiusd.host)

def run_acct(config):
    from twisted.internet import reactor
    from toughradius.radiusdd.server.master import RADIUSMaster
    acct_protocol = RADIUSMaster(config,service='acct')
    reactor.listenUDP(int(config.radiusd.acct_port), acct_protocol, interface=config.radiusd.host)

def run_worker(config,dbengine,**kwargs):
    from twisted.internet import reactor
    from toughradius.manage import settings
    from toughradius.radiusdd.server.master import RADIUSAuthWorker,RADIUSAcctWorker
    _cache = kwargs.pop("cache",CacheManager(settings.redis_conf(config),cache_name='WorkerCache-%s'%os.getpid()))
    # app event init
    if not kwargs.get('standalone'):
        logger.info("start register radiusd events")
        dispatch.register(log_trace.LogTrace(redis_conf(config)),check_exists=True)
        event_params= dict(dbengine=dbengine, mcache=_cache, aes=kwargs.pop('aes',None))
        event_path = os.path.abspath(os.path.dirname(toughradius.events.__file__))
        dispatch.load_events(event_path,"toughradius.events",event_params=event_params)
    logger.info('start radius worker: %s' % RADIUSAuthWorker(config,dbengine,radcache=_cache))
    logger.info('start radius worker: %s' % RADIUSAcctWorker(config,dbengine,radcache=_cache))