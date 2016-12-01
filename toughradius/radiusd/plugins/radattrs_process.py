#!/usr/bin/env python
#coding=utf-8

import traceback
from toughlib import logger, utils


def radius_process(resp=None, resp_attrs={}):
    try:
        if not resp_attrs:
            return resp

        attrs = resp_attrs.get("attrs",{})
        for attr_name, attr_value in attrs.iteritems():
            try:
                # todo: May have a type matching problem
                resp.AddAttribute(utils.safestr(attr_name), attr_value)
            except:
                errmsg = u"add radius attr error %s:%s, %s"%(attr_name,attr_value,traceback.format_exc())
                logger.error(errmsg, trace="radius",tag="radius_attrs_process_error")
        return resp
    except Exception as err:
        logger.exception(err,trace="radius",tag="radius_attrs_process_error")
        return resp

plugin_name = 'radius attrs process'
plugin_types = ['radius_accept']
plugin_priority = 220
plugin_func = radius_process