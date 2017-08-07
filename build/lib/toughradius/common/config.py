#!/usr/bin/env python
#coding=utf-8

import json
import click
import os

ENVKEYS = [
    'CONFDIR'
]

class ConfigDict(dict):

    def __getattr__(self, key):
        try:
            result = self[key]
            if result and isinstance(result, dict):
                result = ConfigDict(result)

            elif result and isinstance(result, (str,unicode)):
                for envkey in ENVKEYS:
                    result = result.replace('{%s}'%envkey,os.environ.get(envkey,""))

                if result.startswith("!include:"):
                    result = Config(result[9:])

            return result
        except KeyError, k:
            return None

    def __setattr__(self, key, value):
        self[key] = value

    def __delattr__(self, key):
        try:
            del self[key]
        except KeyError, k:
            raise AttributeError, k

    def __repr__(self):
        return '<RadiusdConfigDict ' + dict.__repr__(self) + '>'


class Config(ConfigDict):

    def __init__(self, conf_file=None, **kwargs):
        assert(conf_file is not None)
        # print "loading config {0}".format(conf_file)
        if not os.path.exists(conf_file):
            print 'config not exists'
            return
        with open(conf_file) as cf:
            self.update(json.loads(cf.read(),"utf-8"))
        self.update(**kwargs)
        self.conf_file = conf_file

    def save(self):
        print "update config {0}".format(self.conf_file)
        with open(self.conf_file,"w") as cf:
            cf.write(json.dumps(self,ensure_ascii=True,indent=4,sort_keys=False))


    def __repr__(self):
        return '<Config ' + dict.__repr__(self) + '>'



def find_config(conf_file=None):
    return Config(conf_file)


@click.command()
@click.option('-c','--conf', default='/etc/toughradius/radiusd.json', help='toughradius config file')
def chk_cfg(conf):
    try:
        os.environ['CONFDIR'] = os.path.dirname(conf)
        from toughradius.common import config as iconfig
        from pprint import pprint as pp
        config = iconfig.find_config(conf)
        print '%s %s %s' % ('-'*50,conf,'-'*50)
        print json.dumps(config,ensure_ascii=True,indent=4,sort_keys=False)
        print '%s logger %s' % ('-'*50,'-'*50)
        print json.dumps(config.logger,ensure_ascii=True,indent=4,sort_keys=False)
        print '%s clients %s' % ('-'*50,'-'*50)
        print json.dumps(config.clients,ensure_ascii=True,indent=4,sort_keys=False)
        print '%s modules %s' % ('-'*50,'-'*50)
        print json.dumps(config.modules,ensure_ascii=True,indent=4,sort_keys=False)
        print '-' * 110
    except:
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    os.environ['CONFDIR'] = '/Users/wangjuntao/toughstruct/ToughRADIUS/etc'
    cfg = find_config("/Users/wangjuntao/toughstruct/ToughRADIUS/etc/radiusd.json")
    print cfg
    print cfg.logger








