#!/usr/bin/env python
#coding=utf-8

from toughlib import logger

def std_domain(resp, domain=None):
    return resp

def ros_domain(resp, domain=None):
    return resp

def panabit_domain(resp, domain=None):
    return resp

def cisco_domain(resp, domain=None):
    return resp

def radback_domain(resp, domain=None):
    if domain:
        resp['Context-Name'] = str(domain)
    return resp

def h3c_domain(resp, domain=None):
    return resp

def zte_domain(resp, domain=None):
    return resp

def huawei_domain(resp, domain=None):
    if domain:
        resp['Huawei-Domain-Name'] = str(domain)
    return resp


domain_funcs = {
    '0': std_domain,
    '9': cisco_domain,
    '2011': huawei_domain,
    '2352': radback_domain,
    '3902': zte_domain,
    '25506': h3c_domain,
    '14988': ros_domain,
    '39999': panabit_domain,
}

def radius_process(resp=None, resp_attrs={}):
    try:
        domain = resp_attrs.get('domain')
        if not domain:
            return resp

        _vendor = str(resp.vendor_id)
        if _vendor in domain_funcs:
            return domain_funcs[_vendor](resp,domain)
        else:
            return std_domain(resp, domain)
    except Exception as err:
        logger.exception(err,trace="radius",tag="radius_domain_process_error")
        return resp

plugin_name = 'radius domain process'
plugin_types = ['radius_accept']
plugin_priority = 210
plugin_func = radius_process


