#!/usr/bin/env python
#coding=utf-8
import logging

def handle_radius(req,reply):
    try:
        if not 'attrs' in reply:
            return reply

        attrs = reply.get("attrs",{})
        for attr_name, attr_value in attrs.iteritems():
            try:
                # todo: May have a type matching problem
                reply.AddAttribute(str(attr_name), attr_value)
            except:
                raise Exception(u"add radius attr error %s:%s"%(attr_name,attr_value))

        return reply
    except Exception:
        logging.error("radius_attrs_process_error",exc_info=True)
        return reply