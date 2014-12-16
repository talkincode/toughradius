#!/usr/bin/env python
#coding=utf-8
from twisted.python import log
from store import store
from settings import *

def process(req=None,resp=None,user=None):

    if req.vendor_id != '9':
        return

    product = store.get_product(user['product_id'])
    in_rate = product.get('input_rate_code')
    out_rate = product.get('output_rate_code')

    if in_rate and out_rate:
        resp.addAtt('Cisco-AVPair','sub-qos-policy-in=%s'%in_rate)
        resp.addAtt('Cisco-AVPair','sub-qos-policy-out=%s'%out_rate)

    domain = user.get('domain_name')
    if domain:
        resp.addAtt('Cisco-AVPair','addr-pool=%s'%domain)

    if user['ip_address']:
        resp['Framed-IP-Address'] = user['ip_address']

    return resp