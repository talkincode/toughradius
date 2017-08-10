import gevent.monkey
gevent.monkey.patch_all()
import os
import sys
import click
import json
import logging
import logging.config

@click.group()
def cli():
    pass

@click.command()
@click.option('-c','--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
def chkcfg(conf):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from pprint import pprint as pp
        config = iconfig.find_config(conf)
        print '%s %s %s' % ('-'*50,conf,'-'*50)
        print json.dumps(config,ensure_ascii=True,indent=4,sort_keys=False)
        print '%s logger %s' % ('-'*50,'-'*50)
        print json.dumps(config.logger,ensure_ascii=True,indent=4,sort_keys=False)
        print '%s clients %s' % ('-'*50,'-'*50)
        print json.dumps(config.clients,ensure_ascii=True,indent=4,sort_keys=False)
        print '%s modules %s' % ('-'*50,'-'*50)
        print json.dumps(config.modules,ensure_ascii=True,indent=4,sort_keys=False)
        print '-' * 110
    except:
        import traceback
        traceback.print_exc()

@click.command()
@click.option('-c','--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d','--debug', is_flag=True)
@click.option('-auth-port','--auth-port', default=0,type=click.INT,help='auth port')
@click.option('-p','--pool-size', default=0,type=click.INT)
def auth(conf,debug,auth_port,pool_size):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from toughradius.radiusd.server import RudiusAuthServer
        config = iconfig.find_config(conf)

        logging.config.dictConfig(config.logger)

        if debug:
            config.radiusd['debug'] = True
        if auth_port > 0:
            config.radiusd['auth_port'] = auth_port
        if pool_size > 0:
            config.radiusd['pool_size'] = pool_size

        address = (config.radiusd.host,config.radiusd.port)
        server = RudiusAuthServer(address, config)
        logging.info(server)
        server.serve_forever()
    except:
        import traceback
        traceback.print_exc()


@click.command()
@click.option('-c','--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
@click.option('-d','--debug', is_flag=True)
@click.option('-acct-port','--acct-port', default=0,type=click.INT,help='acct port')
@click.option('-p','--pool-size', default=0,type=click.INT)
def acct(conf,debug,acct_port,pool_size):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        config = iconfig.find_config(conf)
        logging.config.dictConfig(config.logger)
        from toughradius.radiusd.server import RudiusAcctServer

        if debug:
            config.radiusd['debug'] = True
        if acct_port > 0:
            config.radiusd['acct_port'] = acct_port
        if pool_size > 0:
            config.radiusd['pool_size'] = pool_size

        address = (config.radiusd.host,config.radiusd.port)
        server = RudiusAcctServer(address, config)
        server.serve_forever()
    except:
        import traceback
        traceback.print_exc()


cli.add_command(chkcfg)
cli.add_command(auth)
cli.add_command(acct)

if __name__ == '__main__':
    cli()



