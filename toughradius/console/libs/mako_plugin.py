#!/usr/bin/env python
#coding=utf-8
import bottle
from bottle import mako_template

if not hasattr(bottle, 'PluginError'):
    class PluginError(bottle.BottleException):
        pass
    bottle.PluginError = PluginError


class MakoPlugin(object):
    name = 'makotpl'
    api = 2

    def __init__(self, sid, lookup, context={}):
        self.name = '%s_%s'%(self.name,sid)
        self.lookup = lookup
        self.context = context
        self.keyword = 'render'

    def render(self, *args, **kwargs):
        kwargs['template_lookup'] = self.lookup
        kwargs['template_settings'] = dict(
            input_encoding='utf-8',
            output_encoding='utf-8',
            encoding_errors='replace'
        )
        kwargs.update(**self.context)
        return mako_template(*args, **kwargs)


    def setup(self, app):
        for other in app.plugins:
            if not isinstance(other, MakoPlugin):
                continue
            if other.name == self.name:
                raise bottle.PluginError("Found another  MakoPlugin plugin with " \
                                         "conflicting settings (non-unique name).")
        if  not self.lookup:
            raise bottle.PluginError('lookup value is None.')

    def apply(self, callback, route):
        def wrapper(*args, **kwargs):
            kwargs['render'] = self.render
            rv = callback(*args, **kwargs)
            return rv
        return wrapper


Plugin = MakoPlugin
