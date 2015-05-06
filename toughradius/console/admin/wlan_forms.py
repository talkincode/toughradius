#coding:utf-8
from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style,input_style

boolean = {0:u"否", 1:u"是"}


param_form = pyforms.Form(
    pyforms.Textbox("wlan_portal_name", description=u"Wlan认证门户标题", **input_style),
    pyforms.Textbox("wlan_free_interval", description=u"免费认证间隔（分钟）", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"Wlan参数配置管理",
    action="/wlan/param/update"
)












