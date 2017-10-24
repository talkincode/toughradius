#!/usr/bin/env python
#coding:utf-8
import gevent.monkey
gevent.monkey.patch_all()
import redis
from gevent import socket
import redis.connection
redis.connection.socket = socket
import os
import time
import click
import json
import logging
import logging.config
import pprint


@click.group()
def cli():
    pass


@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
def chkcfg(conf):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from pprint import pprint as pp
        config = iconfig.find_config(conf)
        print '%s %s %s' % ('-' * 50, conf, '-' * 50)
        print json.dumps(config, ensure_ascii=True, indent=4, sort_keys=False)
        print '%s logger %s' % ('-' * 50, '-' * 50)
        print json.dumps(config.logger, ensure_ascii=True, indent=4, sort_keys=False)
        print '%s clients %s' % ('-' * 50, '-' * 50)
        print json.dumps(config.clients, ensure_ascii=True, indent=4, sort_keys=False)
        print '-' * 110
    except Exception as err:
        click.echo(click.style('chk config error %s' % err.message, fg='red'))


@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d', '--debug', is_flag=True)
@click.option('-auth-port', '--auth-port', default=0, type=click.INT, help='auth port')
@click.option('-p', '--pool-size', default=0, type=click.INT)
def auth(conf, debug, auth_port, pool_size):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from toughradius.radiusd.master import RudiusAuthServer
        config = iconfig.find_config(conf)

        logging.config.dictConfig(config.logger)

        if debug:
            config.radiusd['debug'] = True
        if auth_port > 0:
            config.radiusd['auth_port'] = auth_port
        if pool_size > 0:
            config.radiusd['pool_size'] = pool_size

        os.environ['TOUGHRADIUS_DEBUG_ENABLE'] = str(int(config.radiusd['debug']))
        address = (config.radiusd.host, int(config.radiusd.auth_port))
        server = RudiusAuthServer(address, config)
        logging.info(server)
        server.serve_forever()
    except Exception as err:
        click.echo(click.style('run auth server error %s' % err.message, fg='red'))


@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d', '--debug', is_flag=True)
@click.option('-acct-port', '--acct-port', default=0, type=click.INT, help='acct port')
@click.option('-p', '--pool-size', default=0, type=click.INT)
def acct(conf, debug, acct_port, pool_size):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        logging.config.dictConfig(config.logger)
        from toughradius.radiusd.master import RudiusAcctServer

        if debug:
            config.radiusd['debug'] = True
        if acct_port > 0:
            config.radiusd['acct_port'] = acct_port
        if pool_size > 0:
            config.radiusd['pool_size'] = pool_size

        os.environ['TOUGHRADIUS_DEBUG_ENABLE'] = str(int(config.radiusd['debug']))
        address = (config.radiusd.host, int(config.radiusd.acct_port))
        server = RudiusAcctServer(address, config)
        logging.info(server)
        server.serve_forever()
    except Exception as err:
        click.echo(click.style('run acct server error %s' % err.message, fg='red'))

@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d', '--debug', is_flag=True)
@click.option('-auth-port', '--auth-port', default=1812, type=click.INT, help='auth port')
@click.option('-acct-port', '--acct-port', default=1813, type=click.INT, help='acct port')
@click.option('-api-port', '--api-port', default=1815, type=click.INT, help='api port')
@click.option('-p', '--pool-size', default=0, type=click.INT)
def radiusd(conf, debug, auth_port,acct_port,api_port, pool_size):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from toughradius.radiusd.master import RudiusAuthServer
        from toughradius.radiusd.master import RudiusAcctServer
        from toughradius.radiusd.apiserver import ApiServer
        config = iconfig.find_config(conf)

        logging.config.dictConfig(config.logger)

        if debug:
            config.radiusd['debug'] = True

        if auth_port > 0:
            config.radiusd['auth_port'] = auth_port

        if acct_port > 0:
            config.radiusd['acct_port'] = acct_port

        if api_port > 0:
            config.api['port'] = api_port

        if pool_size > 0:
            config.radiusd['pool_size'] = pool_size

        os.environ['TOUGHRADIUS_DEBUG_ENABLE'] = str(int(config.radiusd['debug']))
        auth_address = (config.radiusd.host, int(config.radiusd.auth_port))
        acct_address = (config.radiusd.host, int(config.radiusd.acct_port))
        auth_server = RudiusAuthServer(auth_address, config)
        acct_server = RudiusAcctServer(acct_address, config)
        auth_server.start()
        gevent.sleep(0.1)
        acct_server.start()
        gevent.sleep(0.1)
        logging.info(auth_server)
        logging.info(acct_server)
        apiserver = ApiServer(config)
        apiserver.start(forever=False)
        gevent.wait()
    except Exception as err:
        click.echo(click.style('run radiusd error %s' % err.message, fg='red'))



@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d', '--debug', is_flag=True)
@click.option('-port', '--port', default=1815, type=click.INT, help='api port')
def apiserv(conf, debug, port):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        logging.config.dictConfig(config.logger)
        from toughradius.radiusd.apiserver import ApiServer

        if debug:
            config.api['debug'] = True
        if port > 0:
            config.api['port'] = port

        apiserver = ApiServer(config)
        apiserver.start(forever=True)
    except Exception as err:
        click.echo(click.style('run apiserver error %s' % err.message, fg='red'))

@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-n','--name', default='toughac')
@click.option('-i','--nasid', default='toughac')
@click.option('-v','--vendor', default='0')
@click.option('-h','--ipaddr', default='127.0.0.1')
@click.option('-s','--secret', default='secret')
@click.option('-coa','--coaport', default=3799,type=click.INT)
def setnas(conf,name,nasid,vendor,ipaddr,secret,coaport):
    try:
        from toughradius.common import rediskeys
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        rediscfg = config.adapters.redis
        rdb = redis.StrictRedis(host=rediscfg.get('host'), port=rediscfg.get("port"), password=rediscfg.get('passwd'))
        with rdb.pipeline() as pipe:
            nas = dict(
                status=1,
                nasid=name,
                name=nasid,
                vendor=vendor,
                ipaddr=ipaddr,
                secret=secret,
                coaport=coaport
            )
            nashkey = rediskeys.NasHKey('toughac','127.0.0.1')
            pipe.hmset(nashkey,nas)
            pipe.zadd(rediskeys.NasSetKey,time.time(), nashkey)
            pipe.execute()
            click.echo(click.style('operate ok: \n%s' % pprint.pformat(nas,indent=4), fg='green'))
    except Exception as err:
        click.echo(click.style('set nas error %s' % err.message, fg='red'))

@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-u', '--username', default='')
@click.option('-p', '--password', default='')
@click.option('--bill-type', default='day', type=click.Choice(['day', 'second','flow']))
@click.option('-input', '--input-rate', default=1048576,type=click.INT)
@click.option('-output', '--output-rate', default=1048576,type=click.INT)
@click.option('--bind-mac', default=0,type=click.INT)
@click.option('--mac-addr', default='')
@click.option('--bind-vlan', default=0,type=click.INT)
@click.option('--time-amount', default=0,type=click.INT)
@click.option('--flow-amount', default=0,type=click.INT)
@click.option('--online-limit', default=0,type=click.INT)
@click.option('--bypass-pwd', default=0,type=click.INT)
@click.option('--expire-date', default='2099-12-30',help='expire date, format:yyyy-mm-dd')
@click.option('--expire-time', default='23:59:59',help='expire time, format:hh:mm:ss')
def setuser(conf,username,password, bill_type,input_rate,output_rate,
            bind_mac,mac_addr,bind_vlan,time_amount,flow_amount,online_limit,bypass_pwd,expire_date,expire_time):
    try:
        username = username.strip()
        password = password.strip()
        if not all([username,password]):
            click.echo(click.style('username and password can not empty', fg='red'))
            return

        from toughradius.common import rediskeys
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        rediscfg = config.adapters.redis
        rdb = redis.StrictRedis(host=rediscfg.get('host'), port=rediscfg.get("port"), password=rediscfg.get('passwd'))
        with rdb.pipeline() as pipe:
            user = dict(
                status=1,
                username=username,
                password=password,
                input_rate=input_rate,
                output_rate=output_rate,
                rate_code='',
                bill_type=bill_type,
                bind_mac=bind_mac,
                bind_vlan=bind_vlan,
                bind_nas=0,
                nas='',
                mac_addr=mac_addr,
                vlanid1=0,
                vlanid2=0,
                time_amount=time_amount,
                flow_amount=flow_amount,
                expire_date=expire_date,
                expire_time=expire_time,
                online_limit=online_limit,
                bypass_pwd=bypass_pwd
            )
            userhkey = rediskeys.UserHKey(username)
            pipe.hmset(userhkey,user)
            pipe.zadd(rediskeys.UserSetKey,time.time(), userhkey)
            pipe.execute()
            click.echo(click.style('operate ok: \n%s' % pprint.pformat(user,indent=4), fg='green'))
    except Exception as err:
        click.echo(click.style('add user error %s' % err.message, fg='red'))

@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-n','--max-count', default=100,type=click.INT)
def userlist(conf,max_count):
    try:
        from toughradius.common import rediskeys
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        rediscfg = config.adapters.redis
        rdb = redis.StrictRedis(host=rediscfg.get('host'), port=rediscfg.get("port"), password=rediscfg.get('passwd'))
        userhkeys = rdb.zrange(rediskeys.UserSetKey,0,max_count-1) or []
        for key in userhkeys:
            click.echo(click.style("%s\n%s\n" % (key, pprint.pformat(rdb.hgetall(key))), fg='green'))
    except Exception as err:
        click.echo(click.style('fetch users error %s' % err.message, fg='blue'))

@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-u', '--username', default='')
def userget(conf,username):
    try:
        from toughradius.common import rediskeys
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        rediscfg = config.adapters.redis
        rdb = redis.StrictRedis(host=rediscfg.get('host'), port=rediscfg.get("port"), password=rediscfg.get('passwd'))
        click.echo(click.style( pprint.pformat(rdb.hgetall(rediskeys.UserHKey(username))), fg='green'))
    except Exception as err:
        click.echo(click.style('get user error %s' % err.message, fg='red'))

@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-n','--max-count', default=100,type=click.INT)
def naslist(conf,max_count):
    try:
        from toughradius.common import rediskeys
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        rediscfg = config.adapters.redis
        rdb = redis.StrictRedis(host=rediscfg.get('host'), port=rediscfg.get("port"), password=rediscfg.get('passwd'))
        userhkeys = rdb.zrange(rediskeys.NasSetKey,0,max_count-1) or []
        for key in userhkeys:
            click.echo(click.style("%s\n%s\n" % (key, pprint.pformat(rdb.hgetall(key))), fg='green'))
    except Exception as err:
        click.echo(click.style('fetch nas list error %s' % err.message, fg='red'))

@click.command()
@click.option('-c', '--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-n','--max-count', default=100,type=click.INT)
@click.option('-u', '--username', default='')
def onlines(conf,max_count,username):
    try:
        from toughradius.common import rediskeys
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        rediscfg = config.adapters.redis
        rdb = redis.StrictRedis(host=rediscfg.get('host'), port=rediscfg.get("port"), password=rediscfg.get('passwd'))
        okeys = rdb.zrange(rediskeys.UserOnlineSetKey(username),0,max_count-1) or []
        click.echo(click.style('total %s' % rdb.zcard(rediskeys.UserOnlineSetKey(username)), fg='blue'))
        for key in okeys:
            click.echo(click.style("%s\n%s\n" % (key, pprint.pformat(rdb.hgetall(key))), fg='green'))
    except Exception as err:
        click.echo(click.style('fetch users error %s' % err.message, fg='red'))

@click.command()
@click.option('-dev', '--develop', is_flag=True)
@click.option('-stable', '--stable', is_flag=True)
def upgrade(develop,stable):
    if develop:
        os.system("pip install -U https://github.com/talkincode/ToughRADIUS/archive/develop.zip")
    elif stable:
        os.system("pip install -U https://github.com/talkincode/ToughRADIUS/archive/master.zip")

cli.add_command(chkcfg)
cli.add_command(auth)
cli.add_command(acct)
cli.add_command(radiusd)
cli.add_command(apiserv)
cli.add_command(upgrade)
cli.add_command(setuser)
cli.add_command(userlist)
cli.add_command(userget)
cli.add_command(naslist)
cli.add_command(onlines)
cli.add_command(setnas)

if __name__ == '__main__':
    cli()
