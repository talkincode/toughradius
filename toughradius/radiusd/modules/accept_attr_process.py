#!/usr/bin/env python
#coding=utf-8
import logging

def handle_radius(req,reply):
    try:
        if not resp_attrs:
            return resp

        attrs = reply.get("attrs",{})
        for attr_name, attr_value in attrs.iteritems():
            try:
                # todo: May have a type matching problem
                reply.AddAttribute(str(attr_name), attr_value)
            except:
                errmsg = u"add radius attr error %s:%s, %s"%(attr_name,attr_value,traceback.format_exc())

        return resp
    except Exception as err:
        logger.exception(err,trace="radius",tag="radius_attrs_process_error")
        return reply