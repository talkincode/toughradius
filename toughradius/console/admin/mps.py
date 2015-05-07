#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from sqlalchemy import func
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.libs.mpsapi import mpsapi
from toughradius.console.base import *
import bottle
import datetime
import json

__prefix__ = "/mps"

app = Bottle()
app.config['__prefix__'] = __prefix__
render = functools.partial(Render.render_app,app)

###############################################################################
# mps manage        
###############################################################################

@app.get('/menus',apply=auth_opr)
def menus(db):   
    menus_data = db.query(models.SlcParam).filter_by(param_name='mps_menus').first()
    if menus_data:
        menus_obj = json.loads(menus_data.param_value) 
    else:
        menus_obj = {}
        menus_obj['button'] = []

    menu_names = {u'精彩推荐': 'menu1', u'产品订购': 'menu2', u'微营业厅': 'menu3'}
    menu_buttons_array = menus_obj['button']

    menudata = {}
    for mbs in menu_buttons_array:
        midx = menu_names[mbs['name']]
        sub_buttons = mbs['sub_button']

        _idx = 1

        for sbmenu in sub_buttons:
            menudata['%s_sub%s_type' % (midx, _idx)] = sbmenu['type']
            menudata['%s_sub%s_name' % (midx, _idx)] = sbmenu['name']
            menudata['%s_sub%s_key' % (midx, _idx)] = sbmenu.get('key', '')
            menudata['%s_sub%s_url' % (midx, _idx)] = sbmenu.get('url', '')
            _idx += 1

    menu_str = json.dumps(menudata, ensure_ascii=False)#.replace('"', '\\"')
    return render("mps_menus", menudata=menu_str)

permit.add_route("%s/menus"%__prefix__,u"公众号菜单",u"微信接入",is_menu=True,order=0)

@app.post('/menus/update',apply=auth_opr)
def mps_menus_update(db):
    def update_menus(menudata):
        ''' 更新远程菜单数据 '''
        _url = mpsapi.wx_sync_menus_url()
        _resp = requests.post(_url, data=menudata)
        _json = _resp.json()
        return _json

    # 更新菜单，保存菜单数据为json字符串 
    menudata = request.params.get("menudata")
    menu_json = json.loads(menudata)

    try:
        menu_object = {'button': []}
        menu_names = {'menu1': u'精彩推荐', 'menu2': u'产品订购', 'menu3': u'微营业厅'}
        for menu in ['menu1', 'menu2', 'menu3']:
            menu_buttons = {'name': menu_names[menu], 'sub_button': []}
            for ms in range(1, 6):
                menu_button = {}
                _menu_type = menu_json['%s_sub%s_type' % (menu, ms)]
                _menu_name = menu_json['%s_sub%s_name' % (menu, ms)]
                _menu_key = menu_json['%s_sub%s_key' % (menu, ms)]
                _menu_url = menu_json['%s_sub%s_url' % (menu, ms)]

                if len(_menu_name) > 1:
                    menu_button['type'] = _menu_type
                    menu_button['name'] = _menu_name
                    if 'click' in _menu_type:
                        menu_button['key'] = _menu_key
                    else:
                        menu_button['url'] = _menu_url

                    menu_buttons['sub_button'].append(menu_button)

            menu_object['button'].append(menu_buttons)

        menu_result = json.dumps(
            menu_object, ensure_ascii=False,
            sort_keys=True,indent=4, 
            separators=(',', ': ')
        )

        menus_param = db.query(models.SlcParam).filter_by(param_name='mps_menus').first()
        if not menus_param:
            menus_param = models.SlcParam()
            menus_param.param_name = 'mps_menus'
            menus_param.param_desc = u"微信公众号菜单数据"
        menus_param.param_value = menu_result
        db.commit()

        _result = update_menus(menu_result.encode('utf-8'))
        if int(_result.get('errcode')) > 0:
            log.err(u"同步菜单失败，" + _result.get('errmsg'))
            return dict(code=0, msg=u'同步微信菜单失败了［%s］，请检查错误再试试' % _result.get('errmsg'))
    except:
        log.err(u"更新菜单失败")
        return dict(code=0, msg=u'更新菜单失败')

    return dict(code=0, msg=u'更新菜单成功')

permit.add_route("%s/menus/update"%__prefix__,u"公众号菜单更新",u"微信接入",is_menu=False,order=0.1)





