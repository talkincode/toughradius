#!/usr/bin/env python
#coding:utf-8
import os
import ConfigParser
from toughradius.tools.shell import shell as sh
from toughradius.tools.secret import gen_secret

def find_config(conf_file=None):
    windows_dir = os.getenv("WINDIR") and os.path.join(os.getenv("WINDIR"),'radiusd.conf') or None
    cfgs = [
        conf_file,
        '/etc/radiusd.conf',
        '/var/toughradius/radiusd.conf',
        './radiusd.conf',
        '~/radiusd.conf',
        windows_dir
    ]
    config = ConfigParser.ConfigParser()
    flag = False
    for c in cfgs:
        if c and os.path.exists(c):
            config.read(c)
            flag = True
            break
   
    if not flag:
        return None
    else:            
        return config
    
    

def setup_config():
    sh.info("set config...")
    config_path = sh.read('set your config file path,[ /etc/radiusd.conf ]') or '/etc/radiusd.conf'
    config = ConfigParser.RawConfigParser()
    sh.info("set default option")
    config.set('DEFAULT', 'debug', (sh.read("set debug [0/1] [0]:") or '0') )
    config.set('DEFAULT', 'tz', (sh.read("time zone [ CST-8 ]:") or 'CST-8') )
    config.set('DEFAULT','secret',gen_secret(32))
    
    sh.info("set database option")
    config.add_section('database')
    config.set('database','dbtype', (sh.read("database type [sqlite]:") or 'sqlite' ))
    config.set('database','dburl',( sh.read("database url [sqlite:////tmp/toughradius.sqlite3]:") or 'sqlite:////tmp/toughradius.sqlite3' ))
    config.set('database','echo',(sh.read("database echo sql [false]:") or 'false' ))
    config.set('database','pool_size',(sh.read("database pool_size [30]:") or '30' ))
    config.set('database','pool_recycle',(sh.read("database pool_recycle(second) [300]:") or '300' ))
    
    sh.info("set radiusd option")
    config.add_section('radiusd')
    config.set('radiusd','authport',(sh.read("radiusd authport [1812]:") or '1812'))
    config.set('radiusd','acctport',(sh.read("radiusd acctport [1813]:") or '1813'))
    config.set('radiusd','adminport',(sh.read("radiusd adminport [1815]:") or '1815'))
    config.set('radiusd','cache_timeout',(sh.read("radiusd cache_timeout (second) [600]:") or '600'))
    config.set('radiusd', 'logfile', (sh.read("log file [ /var/log/radiusd.log ]:") or '/var/log/radiusd.log') )
    
    sh.info("set admin option")
    config.add_section('admin')
    config.set('admin','port',(sh.read("admin http port [1816]:") or '1816'))
    config.set('admin', 'logfile', (sh.read("log file [ /var/log/admin.log ]:") or '/var/log/admin.log') )
    
    sh.info("set customer option")
    config.add_section('customer')
    config.set('customer','port',(sh.read("customer http port [1817]:") or '1817'))
    config.set('customer', 'logfile', (sh.read("log file [ /var/log/customer.log ]:") or '/var/log/customer.log') )
    
    with open(config_path,'wb') as configfile:
        config.write(configfile)
        sh.succ("config save to %s"%config_path)

def echo_radiusd_cnf():
    return '''[DEFAULT]
debug = 0
tz = CST-8
secret = %s

[database]
dbtype = sqlite
dburl = sqlite:////tmp/toughradius.sqlite3
echo = false
# dbtype = mysql
# dburl = mysql://root:root@127.0.0.1/toughradius0?charset=utf8
# pool_size = 120
# pool_recycle = 300

[radiusd]
acctport = 1813
adminport = 1815
authport = 1812
cache_timeout = 600
logfile = /var/log/radiusd.log

[admin]
port = 1816
logfile = /var/log/admin.log

[customer]
port = 1817
logfile = /var/log/customer.log
'''%gen_secret(32)


def echo_app_tac(app):
    if app == 'radiusd':
        return '''from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
from twisted.application import service, internet
from toughradius.tools import config
from toughradius.tools.dbengine import DBEngine
from toughradius.radiusd import server
application = service.Application("ToughRADIUS Radiusd Application")
config = config.find_config()
service = server.run(config,DBEngine(config).get_engine(),True)
service.setServiceParent(application)'''
    elif app == 'admin':
        return '''from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
from twisted.application import service, internet
from toughradius.tools import config
from toughradius.tools.dbengine import DBEngine
from toughradius.console import admin_app
application = service.Application("ToughRADIUS Admin Application")
config = config.find_config()
service = admin_app.run(config,DBEngine(config).get_engine(),True)
service.setServiceParent(application)'''
    elif app == 'customer':
        return '''from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
from twisted.application import service, internet
from toughradius.tools import config
from toughradius.tools.dbengine import DBEngine
from toughradius.console import customer_app
application = service.Application("ToughRADIUS Customer Application")
config = config.find_config()
service = customer_app.run(config,DBEngine(config).get_engine(),True)
service.setServiceParent(application)'''
    elif app == 'standalone':
        return '''from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
from twisted.application import service, internet
from toughradius.tools import config
from toughradius.tools.dbengine import get_engine
from toughradius.console import admin_app
from toughradius.console import customer_app
from toughradius.radiusd import server
application = service.Application("ToughRADIUS Standalone Application")
config = config.find_config()
db_engine = get_engine(config)
service = server.run(config,db_engine,True)
admin_app.run(config,db_engine,True).setServiceParent(service)
customer_app.run(config,db_engine,True).setServiceParent(service)
service.setServiceParent(application)'''