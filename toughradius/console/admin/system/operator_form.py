#!/usr/bin/env python
# coding=utf-8
from toughradius.common import pyforms
from toughradius.common.pyforms import rules
from toughradius.common.pyforms.rules import button_style, input_style

opr_status_dict = {0: u'正常', 1: u"停用"}

def operator_add_form(nodes=[], products=[]):
    return pyforms.Form(
        pyforms.Textbox("operator_name", rules.len_of(2, 32), description=u"操作员名称", required="required", **input_style),
        pyforms.Textbox("operator_desc", rules.len_of(0, 255), description=u"操作员姓名", **input_style),
        pyforms.Password("operator_pass", rules.len_of(6, 128), description=u"操作员密码", required="required",**input_style),
        pyforms.Dropdown("operator_status", description=u"操作员状态", args=opr_status_dict.items(), required="required",**input_style),
        pyforms.Dropdown("operator_nodes", description=u"关联区域(多选)", args=nodes, required="required",multiple="multiple", size=4, **input_style),
        pyforms.Dropdown("operator_products", description=u"关联资费(多选)", args=products, required="required",multiple="multiple", size=6, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"增加操作员",
        action="/operator/add"
    )


def operator_update_form(nodes=[], products=[]):
    return pyforms.Form(
        pyforms.Hidden("id", description=u"编号"),
        pyforms.Textbox("operator_name", description=u"操作员名称", readonly="readonly", **input_style),
        pyforms.Textbox("operator_desc", rules.len_of(0, 255), description=u"操作员姓名", **input_style),
        pyforms.Password("operator_pass", rules.len_of(0, 128), description=u"操作员密码(留空不修改)", autocomplete="off",**input_style),
        pyforms.Dropdown("operator_status", description=u"操作员状态", args=opr_status_dict.items(), required="required",**input_style),
        pyforms.Dropdown("operator_nodes", description=u"关联区域(多选)", args=nodes, required="required",multiple="multiple", size=4, **input_style),
        pyforms.Dropdown("operator_products", description=u"关联资费(多选)", args=products, required="required",multiple="multiple", size=6, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"修改操作员",
        action="/operator/update"
    )