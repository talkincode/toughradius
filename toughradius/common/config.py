#!/usr/bin/env python
#coding=utf-8

import json
import os

class ConfigDict(dict):

    def __getattr__(self, key):
        try:
            result = self[key]
            if result and isinstance(result, dict):
                result = ConfigDict(result)
            return result
        except KeyError, k:
            raise AttributeError, k

    def __setattr__(self, key, value):
        self[key] = value

    def __delattr__(self, key):
        try:
            del self[key]
        except KeyError, k:
            raise AttributeError, k

    def __repr__(self):
        return '<ConfigDict ' + dict.__repr__(self) + '>'


class Config(ConfigDict):

    def __init__(self, conf_file=None, **kwargs):
        assert(conf_file is not None)
        print "loading config {0}".format(conf_file)
        if not os.path.exists(conf_file):
            print 'config not exists'
            return
        with open(conf_file) as cf:
            self.update(json.loads(cf.read()))
        self.update(**kwargs)
        self.conf_file = conf_file

    def save(self):
        print "update config {0}".format(self.conf_file)
        with open(self.conf_file,"w") as cf:
            cf.write(json.dumps(self,ensure_ascii=True,indent=4,sort_keys=True))


    def __repr__(self):
        return '<Config ' + dict.__repr__(self) + '>'


def find_config(conf_file=None):
    return Config(conf_file)

if __name__ == "__main__":
    cfg = find_config("/tmp/tpconfig21")
    print cfg
    admin = {}
    admin['host'] = '192.1.1.1'
    cfg.update(admin=admin)
    cfg.ccc = u'cc'
    cfg.save()
    print cfg







