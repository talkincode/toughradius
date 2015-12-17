#coding=utf-8
from twisted.python import log
from toughradius.radiusd import plugins
import logging

class Middleware():
    _plugin_modules = {}
    _plugin_loaded = False

    @classmethod
    def load_plugins(cls):
        if cls._plugin_loaded:
            return
        for name in plugins.__all__:
            try:
                __import__('toughradius.radiusd.plugins',globals(),locals(), [name])
                cls.add_plugin(name,getattr(plugins,name))
                log.msg('Plugin %s loaded success.' % name,level=logging.INFO)
            except Exception as err:
                log.err(err,'Fail to load plugin %s' % (name),level=logging.INFO)
        cls._plugin_loaded = True

    @classmethod
    def add_plugin(cls,name,plugin):
        if not hasattr(plugin, 'process'):
            log.err('Plugin %s has no method named process, ignore it',level=logging.INFO)
            return False
        cls._plugin_modules[name] = plugin
        return True

    def process(self,name,**kwargs):
        if name not in self._plugin_modules:
            raise Exception('Plugin %s not find. ' % (name))
        try:
            # log.msg('Plugin %s match.'%name,level=logging.DEBUG)
            return self._plugin_modules[name].process(**kwargs)
        except Exception as err:
            log.err(err,'Plugin %s failed to process.' % name,level=logging.INFO)
            import traceback
            traceback.print_exc()
            return False
             

#first init plugins
Middleware.load_plugins()

if __name__ == '__main__':
    pass
