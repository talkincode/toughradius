#coding=utf-8

from twisted.python import log
import plugins

class Middleware():
    _plugin_modules = {}
    _plugin_loaded = False

    @classmethod
    def load_plugins(cls):
        if cls._plugin_loaded:
            return
        for name in plugins.__all__:
            try:
                __import__('plugins.%s' % name)
                cls.add_plugin(name,getattr(plugins, name))
                log.msg('Plugin %s loaded success.' % name)
            except Exception as err:
                log.err(err,'Fail to load plugin %s' % (name))
        cls._plugin_loaded = True

    @classmethod
    def add_plugin(cls,name,plugin):
        if not hasattr(plugin, 'process'):
            log.err('Plugin %s has no method named process, ignore it')
            return False
        cls._plugin_modules[name] = plugin
        return True

    def process(self,name,**kwargs):
        if name not in self._plugin_modules:
            raise Exception('Plugin %s not find. ' % (name))
        try:
            return self._plugin_modules[name].process(**kwargs)
        except Exception as err:
            log.err(err,'Plugin %s failed to process.' % name)
            raise 

#第一次导入此模块时进行初始化加载
Middleware.load_plugins()

if __name__ == '__main__':
    pass
