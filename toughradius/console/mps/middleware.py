#coding=utf-8

import logging
from toughradius.console.mps import plugins
from toughradius.console.mps.plugins import superbot
from twisted.python import log

class MiddleWare(object):

    _plugin_modules = []
    _plugin_loaded = False

    def __init__(self,config=None):
        self.config = config 
        self.load_plugins()

    def load_plugins(self):
        """
        加载所有插件
        """
        if self._plugin_loaded:
            return
        for name in plugins.__all__:
            try:
                __import__('toughradius.console.mps.plugins.%s' % name)
                self.add_plugin(getattr(plugins, name))
                log.msg('Mps plugin %s loaded success.' % name)
            except:
                import traceback
                log.err('Fail to load mps plugin %s' % (name))
                traceback.print_exc()
        self._plugin_loaded = True

    def add_plugin(self, plugin):
        if not hasattr(plugin, 'test'):
            log.err('Mps plugin %s has no method named test, ignore it')
            return False
        if not hasattr(plugin, 'respond'):
            log.err('Mps plugin %s has no method named respond, ignore it')
            return False
        self._plugin_modules.append(plugin)
        return True

    def respond(self, data, msg=None, db=None,**kwargs):
        """
        调用插件进行消息处理，传入参数：
        @data 消息字符串内容
        @msg 原始消息对象
        @db 传入数据库会话，如果有数据库读写
        """
        response = None
        for plugin in self._plugin_modules:
            try:
                if plugin.test(data, msg, db,**kwargs):
                    log.msg('Mps plugin %s is match' % plugin.__name__)
                    response = plugin.respond(data, msg, db,**kwargs)
            except Exception as err:
                log.err('Mps plugin %s failed to respond. %s' %( plugin.__name__,str(err)))
                continue
            if response:
                break

        return response or superbot.respond(data, msg, db,**kwargs) or u''

