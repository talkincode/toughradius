#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys
import os
import time
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from toughlib import utils
from toughradius.manage import models
from toughlib.dbengine import get_engine
from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.manage.settings import FREE_FEE_PID, FreeFee
from hashlib import md5


def inittest(db):
    node = models.TrNode()
    node.id = 100
    node.node_name = u'测试区域'
    node.node_desc = u'测试区域'
    db.add(node)
    
    product1 = models.TrProduct()
    product1.product_name = u"预付费时长每小时2元"
    product1.product_policy = 1
    product1.product_status = 0
    product1.fee_months = 0
    product1.fee_times = 0
    product1.fee_flows = 0
    product1.bind_mac = 0
    product1.bind_vlan = 0
    product1.concur_number = 1
    product1.fee_price = 200
    product1.fee_period =  '' 
    product1.input_max_limit = 1048576 * 2
    product1.output_max_limit = 1048576 * 2
    product1.create_time = utils.get_currtime()
    product1.update_time = utils.get_currtime()
    db.add(product1)


    product2 = models.TrProduct()
    product2.product_name = u"预付费包月30元"
    product2.product_policy = 0
    product2.product_status = 0
    product2.fee_months = 0
    product2.fee_times = 0
    product2.fee_flows = 0
    product2.bind_mac = 0
    product2.bind_vlan = 0
    product2.concur_number = 1
    product2.fee_price = 3000
    product2.fee_period =  '' 
    product2.input_max_limit = 1048576 * 2
    product2.output_max_limit = 1048576 * 2
    product2.create_time = utils.get_currtime()
    product2.update_time = utils.get_currtime()
    db.add(product2)


    product3 = models.TrProduct()
    product3.product_name = u"买断包月12个月500元"
    product3.product_policy = 2
    product3.product_status = 0
    product3.fee_months = 12
    product3.fee_times = 0
    product3.fee_flows = 0
    product3.bind_mac = 0
    product3.bind_vlan = 0
    product3.concur_number = 1
    product3.fee_price = 50000
    product3.fee_period =  '' 
    product3.input_max_limit = 1048576 * 2
    product3.output_max_limit = 1048576 * 2
    product3.create_time = utils.get_currtime()
    product3.update_time = utils.get_currtime()
    db.add(product3)


    product4 = models.TrProduct()
    product4.product_name = u"买断时长100元50小时"
    product4.product_policy = 3
    product4.product_status = 0
    product4.fee_months = 0
    product4.fee_times = 50 * 3600
    product4.fee_flows = 0
    product4.bind_mac = 0
    product4.bind_vlan = 0
    product4.concur_number = 1
    product4.fee_price = 10000
    product4.fee_period =  '' 
    product4.input_max_limit = 1048576 * 2
    product4.output_max_limit = 1048576 * 2
    product4.create_time = utils.get_currtime()
    product4.update_time = utils.get_currtime()
    db.add(product4)


    product5 = models.TrProduct()
    product5.product_name = u"预付费流量每MB0.05元"
    product5.product_policy = 4
    product5.product_status = 0
    product5.fee_months = 0
    product5.fee_times = 0
    product5.fee_flows = 0
    product5.bind_mac = 0
    product5.bind_vlan = 0
    product5.concur_number = 1
    product5.fee_price = 5
    product5.fee_period =  '' 
    product5.input_max_limit = 1048576 * 2
    product5.output_max_limit = 1048576 * 2
    product5.create_time = utils.get_currtime()
    product5.update_time = utils.get_currtime()
    db.add(product5)


    product6 = models.TrProduct()
    product6.product_name = u"买断流量5元100MB"
    product6.product_policy = 5
    product6.product_status = 0
    product6.fee_months = 0
    product6.fee_times = 0
    product6.fee_flows = 100
    product6.bind_mac = 0
    product6.bind_vlan = 0
    product6.concur_number = 1
    product6.fee_price = 500
    product6.fee_period =  '' 
    product6.input_max_limit = 1048576 * 2
    product6.output_max_limit = 1048576 * 2
    product6.create_time = utils.get_currtime()
    product6.update_time = utils.get_currtime()
    db.add(product6)

    db.commit()
    db.close()


def update(config,force=False):
    try:
        db_engine = get_engine(config)
        print 'init test datat'
        db = scoped_session(sessionmaker(bind=db_engine, autocommit=False, autoflush=True))()
        inittest(db)
    except:
        import traceback
        traceback.print_exc()



