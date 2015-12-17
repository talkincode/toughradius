#!/usr/bin/env python
#coding=utf-8

from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}

node_add_form = pyforms.Form(
    pyforms.Textbox("node_name", rules.len_of(2, 32), description=u"区域名称", required="required", **input_style),
    pyforms.Textbox("node_desc", rules.len_of(2, 128), description=u"区域描述", required="required", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加区域",
    action="/node/add"
)

node_update_form = pyforms.Form(
    pyforms.Hidden("id", description=u"区域编号"),
    pyforms.Textbox("node_name", rules.len_of(2, 32), description=u"区域名称", required="required", **input_style),
    pyforms.Textbox("node_desc", rules.len_of(2, 128), description=u"区域描述", required="required", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改区域",
    action="/node/update"
)
