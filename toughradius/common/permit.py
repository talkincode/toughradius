#!/usr/bin/env python
# coding=utf-8
import time
import os
import importlib
import logging
from cyclone.web import RequestHandler
from cyclone.websocket import WebSocketHandler


class Permit():
    """ 权限菜单管理
    """
    routes = {}
    all_handlers = []

    def add_route(self, handle_cls, path, name, category, handle_params={}, is_menu=False, order=time.time(),
                  is_open=True):
        """ 注册权限
        """
        if not path: return
        self.routes[path] = dict(
            path=path,  # 权限url路径
            name=name,  # 权限名称
            category=category,  # 权限目录
            is_menu=is_menu,  # 是否在边栏显示为菜单
            oprs=[],  # 关联的操作员
            order=order,  # 排序
            is_open=is_open  # 是否开放授权
        )
        self.add_handler(handle_cls, path, handle_params)

    def add_handler(self, handle_cls, path, handle_params={}):
        print ("add handler [%s:%s]" % (path, repr(handle_cls)))
        self.all_handlers.append((path, handle_cls, handle_params))

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
        if not path or not opr:
            return False
        if path not in self.routes:
            return False
        return opr in self.routes[path]['oprs']

    def route(self, url_pattern, menuname=None, category=None, is_menu=False, order=0, is_open=True):
        selfobj = self

        def handler_wapper(cls):
            assert (issubclass(cls, RequestHandler) or issubclass(cls, WebSocketHandler))
            if not menuname:
                self.add_handler(cls, url_pattern)
            else:
                selfobj.add_route(cls, url_pattern, menuname, category, order=order, is_menu=is_menu, is_open=is_open)
            return cls

        return handler_wapper


# 全局实例
permit = Permit()


def load_handlers(handler_path=None, pkg_prefix=None, excludes=('__init__', 'base', '.svn', '.DS_Store', 'views')):
    hds = set(os.path.splitext(it)[0] for it in os.listdir(handler_path))
    hds = [it for it in hds if it not in excludes]
    for hd in hds:
        try:
            sub_module = os.path.join(handler_path, hd)
            if os.path.isdir(sub_module):
                logging.info('load sub module %s' % hd)
                load_handlers(
                    handler_path=sub_module,
                    pkg_prefix="{0}.{1}".format(pkg_prefix, hd)
                )

            _hd = "{0}.{1}".format(pkg_prefix, hd)
            logging.info('load_module %s' % _hd)
            importlib.import_module(_hd)
        except:
            logging.exception("load_module error ")
            continue



