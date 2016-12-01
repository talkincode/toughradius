#!/usr/bin/env python
#coding=utf-8
from toughradius.common import btforms
from toughradius.common.btforms import dataform
from toughradius.common.btforms import rules
from toughradius.common.btforms.rules import button_style, input_style
button_style ={"class":"btn btn-md bg-navy"}

boolean = {0: u"否", 1: u"是"}
timetype = {0: u"标准时区,北京时间", 1: u"时区和时间同区"}
bastype = {
    '0': u'标准',
    '3041': u'阿尔卡特',
    '2352': u'爱立信',
    '2011': u'华为',
    '25506': u'H3C',
    '3902': u'中兴',
    '10055': u'爱快',
    '14988': u'RouterOS'
}

bas_add_form = btforms.Form(
    btforms.Textbox("ip_addr", rules.is_ip, description=u"接入设备地址",  **input_style),
    btforms.Textbox("dns_name", rules.len_of(1, 128), description=u"DNS域名", help=u"动态IP专用", **input_style),
    btforms.Textbox("bas_name", rules.len_of(2, 64), description=u"接入设备名称", required="required", **input_style),
    btforms.Textbox("bas_secret", rules.is_alphanum2(4, 32), description=u"共享秘钥", required="required", **input_style),
    btforms.Dropdown("vendor_id", description=u"接入设备类型", args=bastype.items(), required="required", **input_style),
    btforms.Textbox("coa_port", rules.is_number, description=u"授权端口", default=3799, required="required",**input_style),
    btforms.Dropdown("time_type", description=u"时间类型", args=timetype.items(), required="required", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加接入设备",
    action="/admin/bas/add"
)

bas_update_form = btforms.Form(
    btforms.Hidden("id", description=u"编号"),
    btforms.Textbox("dns_name", rules.len_of(1, 128), description=u"DNS域名", help=u"动态IP专用", **input_style),
    btforms.Textbox("ip_addr", rules.is_ip, description=u"接入设备地址",  **input_style),
    btforms.Textbox("bas_name", rules.len_of(2, 64), description=u"接入设备名称", required="required", **input_style),
    btforms.Textbox("bas_secret", rules.is_alphanum2(4, 32), description=u"共享秘钥", required="required", **input_style),
    btforms.Dropdown("vendor_id", description=u"接入设备类型", args=bastype.items(), required="required", **input_style),
    btforms.Textbox("coa_port", rules.is_number, description=u"授权端口", default=3799, required="required",**input_style),
    btforms.Dropdown("time_type", description=u"时间类型", args=timetype.items(), required="required", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改接入设备",
    action="/admin/bas/update"
)