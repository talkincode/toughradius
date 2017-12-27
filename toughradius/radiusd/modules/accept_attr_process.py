#!/usr/bin/env python
#coding=utf-8
import logging

logger = logging.getLogger(__name__)

def handle_radius(req,reply):
    try:
        if 'radius_attrs' not in reply:
            return reply

        radius_attrs = reply.get("radius_attrs",{})
        for attr_name, attr_value in radius_attrs.iteritems():
            try:
                # todo: May have a type matching problem
                reply.AddAttribute(str(attr_name), attr_value)
            except:
                raise Exception(u"add radius attr error %s:%s"%(attr_name,attr_value))

        return reply
    except Exception:
        logger.error("radius_attrs_process_error",exc_info=True)
        return reply