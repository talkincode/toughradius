#!/usr/bin/env python
#coding=utf-8

from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}
roster_type = {0: u"白名单", 1: u"黑名单"}

roster_add_form = pyforms.Form(
    pyforms.Textbox("mac_addr", description=u"MAC地址", required="required", **input_style),
    pyforms.Textbox("begin_time", description=u"开始时间", required="required", **input_style),
    pyforms.Textbox("end_time", description=u"结束时间", required="required", **input_style),
    pyforms.Dropdown("roster_type", args=roster_type.items(), description=u"类型", value=0, required="required",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加黑白名单",
    action="/roster/add"
)

roster_update_form = pyforms.Form(
    pyforms.Hidden("id", description=u"编号"),
    pyforms.Textbox("mac_addr", description=u"MAC地址", readonly="readonly", **input_style),
    pyforms.Textbox("begin_time", description=u"开始时间", required="required", **input_style),
    pyforms.Textbox("end_time", description=u"结束时间", required="required", **input_style),
    pyforms.Dropdown("roster_type", args=roster_type.items(), description=u"类型", value=0, required="required",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改黑白名单",
    action="/roster/update"
)
