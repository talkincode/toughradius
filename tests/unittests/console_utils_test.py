#!/usr/bin/env python
#coding:utf-8
from toughradius.console.libs import utils
# import ConfigParser
# import os
# from toughradius.console.admin.admin import app as mainapp
# from toughradius.console.admin.ops import app as ops_app
# from toughradius.console.admin.business import app as bus_app
# from toughradius.console.admin.card import app as card_app
# from toughradius.console.admin.product import app as product_app

def test_mb2kb2mb():
    assert utils.mb2kb(0) == 0 
    assert utils.mb2kb('') == 0 
    assert utils.mb2kb(None) == 0 
    assert utils.kb2mb(0) == '0.00'
    assert utils.kb2mb(None) == '0.00'
    assert utils.kb2mb('') == '0.00'
    
# def test_admin_server():
#     from toughradius.console.admin_app import AdminServer
#     config = ConfigParser.ConfigParser()
#     config.read('%s/radiusd.conf'%os.path.dirname(__file__))
#     app = mainapp
#     subapps = [ops_app,bus_app,card_app,product_app]
#     admin = AdminServer(config,app=app,subapps=subapps)
#     assert admin.use_ssl is not None
    