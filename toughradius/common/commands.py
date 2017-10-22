#!/usr/bin/env python
#coding:utf-8
import gevent.monkey
gevent.monkey.patch_all()
import redis
from gevent import socket
import redis.connection
redis.connection.socket = socket
import os
import re
import sys
import signal
import click
import json
import logging
import logging.config


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
    except:
        import traceback
        traceback.print_exc()


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
    except:
        import traceback
        traceback.print_exc()


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
    except:
        import traceback
        traceback.print_exc()

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
        # gevent.signal(signal.SIGTERM, auth_server.close)
        # gevent.signal(signal.SIGINT, auth_server.close)
        # gevent.signal(signal.SIGTERM, acct_server.close)
        # gevent.signal(signal.SIGINT, acct_server.close)
        auth_server.start()
        gevent.sleep(0.1)
        acct_server.start()
        gevent.sleep(0.1)
        logging.info(auth_server)
        logging.info(acct_server)
        apiserver = ApiServer(config)
        apiserver.start(forever=False)
        gevent.wait()
    except:
        import traceback
        traceback.print_exc()



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
    except:
        import traceback
        traceback.print_exc()

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

if __name__ == '__main__':
    cli()
