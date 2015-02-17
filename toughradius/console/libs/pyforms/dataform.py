#!/usr/bin/env python
#coding=utf-8
import copy
from toughradius.console.libs.pyforms import  attrget
from toughradius.console.libs.pyforms import storage
from toughradius.console.libs.pyforms import AttributeList

class Form(object):

    def __init__(self, *inputs, **kw):
        self.inputs = inputs
        self.valid = True
        self.note = None
        self.validators = kw.pop('validators', [])

    def __call__(self, x=None):
        o = copy.deepcopy(self)
        if x: o.validates(x)
        return o

    @property
    def errors(self):
        return ",".join([u"%s error,%s" % (i.description, i.note) for i in self.inputs if i.note])

    def validates(self, source=None, _validate=True, **kw):
        source = source or kw
        out = True
        for i in self.inputs:
            v = attrget(source, i.name)
            if _validate:
                out = i.validate(v) and out
            else:
                i.set_value(v)
        if _validate:
            out = out and self._validate(source)
            self.valid = out
        return out

    def _validate(self, value):
        self.value = value
        for v in self.validators:
            if not v.valid(value):
                self.note = v.msg
                return False
        return True

    def fill(self, source=None, **kw):
        return self.validates(source, _validate=False, **kw)

    def __getitem__(self, i):
        for x in self.inputs:
            if x.name == i: return x
        raise KeyError, i

    def __getattr__(self, name):
        # don't interfere with deepcopy
        inputs = self.__dict__.get('inputs') or []
        for x in inputs:
            if x.name == name: return x
        raise AttributeError, name

    def get(self, i, default=None):
        try:
            return self[i]
        except KeyError:
            return default

    def _get_d(self):  #@@ should really be form.attr, no?
        return storage([(i.name, i.get_value()) for i in self.inputs])

    d = property(_get_d)


class Item(object):
    def __init__(self, name, *validators, **attrs):
        self.name = name
        self.validators = validators
        self.attrs = attrs = AttributeList(attrs)

        self.description = attrs.pop('description', name)
        self.value = attrs.pop('value', None)
        self.note = None

        self.id = attrs.setdefault('id', self.get_default_id())


    def get_default_id(self):
        return self.name

    def validate(self, value):
        self.set_value(value)

        for v in self.validators:
            if not v.valid(value):
                self.note = v.msg
                return False
        return True

    def set_value(self, value):
        self.value = value

    def get_value(self):
        return self.value

    def addatts(self):
        # add leading space for backward-compatibility
        return " " + str(self.attrs)