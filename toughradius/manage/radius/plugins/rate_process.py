#!/usr/bin/env python
#coding=utf-8


def std_rate(resp, _in, _out):
    # input_limit = str(_in)
    # output_limit = str(_out)
    # _class = input_limit.zfill(8) + input_limit.zfill(8) + output_limit.zfill(8) + output_limit.zfill(8)
    # resp['Class'] = _class
    return resp


def ros_rate(resp, _in, _out):
    _irate = _in / 1024
    _orate = _out / 1024
    resp['Mikrotik-Rate-Limit'] = '%sk/%sk' % (_irate, _orate)
    return resp


def aikuai_rate(resp, _in, _out):
    _irate = _in / 1024 / 8
    _orate = _out / 1024 / 8
    resp['RP-Upstream-Speed-Limit'] = _irate
    resp['RP-Downstream-Speed-Limit'] = _orate
    return resp


def cisco_rate(resp, _in, _out):
    return resp


def radback_rate(resp, _in, _out):
    return resp


def h3c_rate(resp, _in, _out):
    resp = std_rate(resp, _in, _out)
    resp['H3C-Input-Average-Rate'] = _in
    resp['H3C-Input-Peak-Rate'] = _in
    resp['H3C-Output-Average-Rate'] = _out
    resp['H3C-Output-Peak-Rate'] = _out
    return resp


def zte_rate(resp, _in, _out):
    resp['ZTE-Rate-Ctrl-Scr-Up'] = _in 
    resp['ZTE-Rate-Ctrl-Scr-Down'] = _out 
    return resp


def huawei_rate(resp, _in, _out):
    input_limit = str(_in)
    output_limit = str(_out)
    _class = input_limit.zfill(8) + input_limit.zfill(8) + output_limit.zfill(8) + output_limit.zfill(8)
    resp['Class'] = _class
    resp['Huawei-Input-Average-Rate'] = _in
    resp['Huawei-Input-Peak-Rate'] = _in
    resp['Huawei-Output-Average-Rate'] = _out
    resp['Huawei-Output-Peak-Rate'] = _out
    return resp


rate_funcs = {
    '0': std_rate,
    '9': cisco_rate,
    '2011': huawei_rate,
    '2352': radback_rate,
    '3902': zte_rate,
    '25506': h3c_rate,
    '14988': ros_rate,
    '10055': aikuai_rate
}


def process(resp=None, input_rate=0, output_rate=0):
    if input_rate == 0 and output_rate == 0:
        return std_rate(resp, 0, 0)

    _vendor = resp.vendor_id or '0'
    return rate_funcs[_vendor](resp, input_rate, output_rate)