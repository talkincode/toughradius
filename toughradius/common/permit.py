#!/usr/bin/env python
# coding=utf-8
import time
import os
import importlib
from toughradius.common import dispatch,logger
import urlparse

class Permit():
    """ 权限菜单管理
    """

    opr_cache = {}

    def __init__(self, parent=None):
        
        if parent:
            self.routes = parent.routes
            self.handlers = parent.handlers
            self.free_routes = parent.free_routes
        else:
            self.routes = {}
            self.handlers = {}
            self.free_routes = []

    def fork(self,opr_name, opr_type=0,rules=[]):
        p = Permit.opr_cache.setdefault(opr_name,Permit(self))
        if opr_type == 0:
            p.bind_super(opr_name)
        else:    
            p.unbind_opr(opr_name)
            for path in rules:
                p.bind_opr(opr_name, path)
        return p


    def add_route(self, handle_cls, path, name, category, 
                  handle_params={}, is_menu=False, 
                  order=time.time(),is_open=True, oem=False,**kwargs):
        """ 注册权限
        """
        if not path: return
        if path in self.routes:
            if self.routes[path].get('oem'):
                return

        self.routes[path] = dict(
            path=path,  # 权限url路径
            name=name,  # 权限名称
            category=category,  # 权限目录
            is_menu=is_menu,  # 是否在边栏显示为菜单
            oprs=[],  # 关联的操作员
            order=order,  # 排序
            is_open=is_open,  # 是否开放授权
            oem=oem #是否定制功能
        )
        self.routes[path].update(**kwargs)
        self.add_handler(handle_cls, path, handle_params)

    def add_handler(self, handle_cls, path, handle_params={}):
        self.handlers[path] = (path, handle_cls, handle_params)

    @property
    def all_handlers(self):
        return self.handlers.values()

    def get_route(self, path):
        """ 获取一个权限资源
        """
        return self.routes.get(path)

    def bind_super(self, opr):
        """ 为超级管理员授权所有权限
        """
        for path in self.routes:
            route = self.routes.get(path)
            route['oprs'].append(opr)

    def bind_opr(self, opr, path):
        """ 为操作员授权
        """
        if not path or path not in self.routes:
            return
        oprs = self.routes[path]['oprs']
        if opr not in oprs:
            oprs.append(opr)

    def unbind_opr(self, opr, path=None):
        """ 接触操作员与权限关联
        """
        if path:
            self.routes[path]['oprs'].remove(opr)
        else:
            for path in self.routes:
                route = self.routes.get(path)
                if route and opr in route['oprs']:
                    route['oprs'].remove(opr)

    def check_open(self, path):
        """ 检查权限是否开放授权
        """
        route = self.routes[path]
        return 'is_open' in route and route['is_open']

    def check_opr_category(self, opr, category):
        """ 检查权限是否在指定目录下
        """
        for path in self.routes:
            route = self.routes[path]
            if opr in route['oprs'] and route['category'] == category:
                return True
        return False

    def build_menus(self, order_cats=[]):
        """ 生成全局内存菜单"""
        menus = [{'category': _cat, 'items': []} for _cat in order_cats]
        for path in self.routes:
            route = self.routes[path]
            for menu in menus:
                if route['category'] == menu['category']:
                    menu['items'].append(route)
        return menus

    def match(self, opr, path):
        """ 检查操作员是否匹配资源
        """
        _url = urlparse.urlparse(path)
        if not _url.path or not opr:
            return False
        if _url.path in self.free_routes:
            return True
        if _url.path not in self.routes:
            return False
        return opr in self.routes[_url.path]['oprs']

    def suproute(self, url_pattern, menuname=None, category=None, 
                  is_menu=False, order=0, is_open=True,oem=False,**kwargs):
        selfobj = self
        def handler_wapper(cls):
            selfobj.add_route(cls, url_pattern, menuname, category, 
                    order=order, is_menu=is_menu, is_open=is_open,oem=oem, admin=True,**kwargs)
            logger.info("add super managed route [%s : %s]" % (url_pattern, repr(cls)))
            return cls

        return handler_wapper

    def route(self, url_pattern, menuname=None, category=None, 
              is_menu=False, order=0, is_open=True,oem=False,**kwargs):
        selfobj = self
        def handler_wapper(cls):
            if not menuname:
                self.add_handler(cls, url_pattern)
                selfobj.free_routes.append(url_pattern)
                logger.info("add free route [%s : %s]" % (url_pattern, repr(cls)))
            else:
                selfobj.add_route(cls, url_pattern, menuname, category, 
                        order=order, is_menu=is_menu, is_open=is_open,oem=oem,**kwargs)
                logger.info("add managed route [%s : %s]" % (url_pattern, repr(cls)))
            return cls

        return handler_wapper


# 全局实例
permit = Permit()


def load_handlers(handler_path=None, pkg_prefix=None, excludes=[]):
    _excludes = ['__init__', 'base', '.svn', '.DS_Store', 'views'] + excludes
    hds = set(os.path.splitext(it)[0] for it in os.listdir(handler_path))
    hds = [it for it in hds if it not in _excludes]
    for hd in hds:
        try:
            sub_module = os.path.join(handler_path, hd)
            if os.path.isdir(sub_module):
                # logger.info('load sub module %s' % hd)
                load_handlers(
                    handler_path=sub_module,
                    pkg_prefix="{0}.{1}".format(pkg_prefix, hd),
                    excludes=excludes
                )

            _hd = "{0}.{1}".format(pkg_prefix, hd)
            # logger.info('load_module %s' % _hd)
            importlib.import_module(_hd)
        except Exception as err:
            logger.error("%s, skip module %s.%s" % (str(err),pkg_prefix,hd))
            import traceback
            traceback.print_exc()
            continue


def load_events(event_path=None,pkg_prefix=None,excludes=[],event_params={}):
    _excludes = ['__init__','settings'] + excludes
    evs = set(os.path.splitext(it)[0] for it in os.listdir(event_path))
    evs = [it for it in evs if it not in _excludes]
    for ev in evs:
        try:
            sub_module = os.path.join(event_path, ev)
            if os.path.isdir(sub_module):
                # logger.info('load sub event %s' % ev)
                load_events(
                    event_path=sub_module,
                    pkg_prefix="{0}.{1}".format(pkg_prefix, ev),
                    excludes=excludes,
                    event_params=event_params,
                )
            _ev = "{0}.{1}".format(pkg_prefix, ev)
            # logger.info('load_event %s with params:%s' % (_ev,repr(event_params)))
            robj = importlib.import_module(_ev)
            if hasattr(robj, 'evobj'):
                dispatch.register(robj.evobj)
            if hasattr(robj, '__call__'):
                dispatch.register(robj.__call__(**event_params))
        except Exception as err:
            logger.error("%s, skip module %s.%s" % (str(err),pkg_prefix,ev))
            import traceback
            traceback.print_exc()
            continue

