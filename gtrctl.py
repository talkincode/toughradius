import os,sys
RUNDIR = os.path.dirname(__file__)
sys.path.insert(0,RUNDIR)
import click
import logging
import logging.config

@click.group()
def cli():
    pass


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


cli.add_command(auth)
cli.add_command(acct)

if __name__ == '__main__':
    cli()



