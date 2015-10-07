#!/usr/bin/env python
#coding:utf-8
import os
import ConfigParser
from toughradius.tools.shell import shell as sh
from toughradius.tools.secret import gen_secret

def find_config(conf_file=None):
    cfgs = [
        conf_file,
        '/etc/radiusd.conf'
    ]
    config = ConfigParser.ConfigParser()
    flag = False
    for c in cfgs:
        if c and os.path.exists(c):
            config.read(c)
            config.set('DEFAULT', 'cfgfile', c)
            sh.info("use config:%s"%c)  
            flag = True
            break
   
    if not flag:
        return None
    else:    
        return config
    


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
    elif app == 'control':
        return '''from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
from twisted.application import service, internet
from toughradius.tools import config
from toughradius.console import control_app
application = service.Application("ToughRADIUS Control Application")
config = config.find_config()
service = control_app.run(config,True)
service.setServiceParent(application)'''
    elif app == 'standalone':
        return '''from autobahn.twisted import choosereactor
choosereactor.install_optimal_reactor(True)
from twisted.application import service, internet
from toughradius.tools import config
from toughradius.tools.dbengine import get_engine
from toughradius.console import admin_app
from toughradius.console import customer_app
from toughradius.console import control_app
from toughradius.radiusd import server
from toughradius.wlan import server as pserver
application = service.Application("ToughRADIUS Standalone Application")
config = config.find_config()
db_engine = get_engine(config)
service = server.run(config,db_engine,True)
admin_app.run(config,db_engine,True).setServiceParent(service)
customer_app.run(config,db_engine,True).setServiceParent(service)
control_app.run(config,True).setServiceParent(service)
pserver.run(config,True).setServiceParent(service)
service.setServiceParent(application)'''