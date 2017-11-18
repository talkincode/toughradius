#!/usr/bin/env python
#coding=utf-8

import logging

logger = logging.getLogger(__name__)

def std_rate(resp, _in, _out, rate_code=None):
    return resp


def ros_rate(resp, _in, _out, rate_code=None):
    _irate = _in / 1024
    _orate = _out / 1024
    resp['Mikrotik-Rate-Limit'] = '%sk/%sk' % (_irate, _orate)
    return resp


def aikuai_rate(resp, _in, _out, rate_code=None):
    _irate = _in / 1024 / 8
    _orate = _out / 1024 / 8
    resp['RP-Upstream-Speed-Limit'] = _irate
    resp['RP-Downstream-Speed-Limit'] = _orate
    return resp


def cisco_rate(resp, _in, _out, rate_code=None):
    return resp


def radback_rate(resp, _in, _out, rate_code=None):
    if rate_code:
        resp['Sub-Profile-Name'] = str(rate_code)
    return resp


def h3c_rate(resp, _in, _out, rate_code=None):
    resp = std_rate(resp, _in, _out)
    resp['H3C-Input-Average-Rate'] = _in
    resp['H3C-Input-Peak-Rate'] = _in
    resp['H3C-Output-Average-Rate'] = _out
    resp['H3C-Output-Peak-Rate'] = _out
    return resp


def zte_rate(resp, _in, _out, rate_code=None):
    resp['ZTE-Rate-Ctrl-Scr-Up'] = _in 
    resp['ZTE-Rate-Ctrl-Scr-Down'] = _out 
    return resp


def huawei_rate(resp, _in, _out, rate_code=None):
    input_limit = str(_in)
    output_limit = str(_out)
    _class = input_limit.zfill(8) + input_limit.zfill(8) + output_limit.zfill(8) + output_limit.zfill(8)
    resp['Class'] = _class      
    resp['Huawei-Input-Average-Rate'] = _in
    resp['Huawei-Input-Peak-Rate'] = _in
    resp['Huawei-Output-Average-Rate'] = _out
    resp['Huawei-Output-Peak-Rate'] = _out
    return resp

def huawei_e1_rate(resp, _in, _out, rate_code=None):
    resp['Huawei-Input-Average-Rate'] = _in / 1024
    resp['Huawei-Input-Peak-Rate'] = _in / 1024
    resp['Huawei-Output-Average-Rate'] = _out / 1024    
    resp['Huawei-Output-Peak-Rate'] = _out / 1024    
    return resp


rate_funcs = {
    '0': std_rate,
    '9': cisco_rate,
    '2011': huawei_rate,
    '2352': radback_rate,
    '3902': zte_rate,
    '25506': h3c_rate,
    '14988': ros_rate,
    '10055': aikuai_rate,
    '30001': huawei_e1_rate
}

def handle_radius(req, reply):
    try:
        logger.debug("accept_rate_process")
        input_rate=int(reply.resp_attrs.get('input_rate',0))
        output_rate=int(reply.resp_attrs.get('output_rate',0))
        rate_code=reply.resp_attrs.get('rate_code')
        if input_rate == 0 and output_rate == 0 and rate_code is None:
            return reply

        _vendor = str(reply.vendor_id)
        if _vendor in rate_funcs:
            return rate_funcs[_vendor](reply, input_rate, output_rate, rate_code)
        else:
            return std_rate(reply, input_rate, output_rate, rate_code)
    except Exception as err:
        logger.exception(err)
        return reply

