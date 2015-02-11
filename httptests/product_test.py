#!/usr/bin/env python
#coding:utf-8

import session

products = [
    dict(
        product_name = u"预付费包月30元",product_policy = 0,
        fee_months = 0,fee_times = 0,
        fee_flows = 0,fee_price = "30.00",
        fee_period = '',concur_number = 1,
        bind_mac = 0,bind_vlan = 0,
        input_max_limit = 1048576,output_max_limit = 1048576,
        product_status = 0
    ),
    dict(
        product_name = u"预付费时长每小时2元",product_policy = 1,
        fee_months = 0,fee_times = 0,
        fee_flows = 0,fee_price = "2.00",
        fee_period = '',concur_number = 1,
        bind_mac = 0,bind_vlan = 0,
        input_max_limit = 1048576,output_max_limit = 1048576,
        product_status = 0
    ),
    dict(
        product_name = u"买断包月12个月500元",product_policy = 2,
        fee_months = 12,fee_times = 0,
        fee_flows = 0,fee_price = "500.00",
        fee_period = '',concur_number = 1,
        bind_mac = 0,bind_vlan = 0,
        input_max_limit = 1048576,output_max_limit = 1048576,
        product_status = 0
    ),
    dict(
        product_name = u"买断时长100元50小时",product_policy = 3,
        fee_months = 0,fee_times = 50,
        fee_flows = 0,fee_price = "100.00",
        fee_period = '',concur_number = 1,
        bind_mac = 0,bind_vlan = 0,
        input_max_limit = 1048576,output_max_limit = 1048576,
        product_status = 0
    ),
    dict(
        product_name = u"预付费流量每MB0.05元",product_policy = 4,
        fee_months = 0,fee_times = 0,
        fee_flows = 0,fee_price = '0.05',
        fee_period = '',concur_number = 1,
        bind_mac = 0,bind_vlan = 0,
        input_max_limit = 1048576,output_max_limit = 1048576,
        product_status = 0
    ),
    dict(
        product_name = u"买断流量5元100MB",product_policy = 5,
        fee_months = 0,fee_times = 0,
        fee_flows = 100,fee_price = '5.00',
        fee_period = '',concur_number = 1,
        bind_mac = 0,bind_vlan = 0,
        input_max_limit = 1048576,output_max_limit = 1048576,
        product_status = 0
    ) 
]

def test_post_product():
    req = session.login()    
    for p in products:
        print 'post product',p
        r = req.post(session.sub_path("/product/add"),data=p)
        assert r.status_code == 200
        
if __name__ == '__main__':
    test_post_product()
