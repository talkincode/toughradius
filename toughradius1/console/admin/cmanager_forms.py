#coding:utf-8
from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style,input_style

def cmanage_add_form(oprs=[]):
    return pyforms.Form(
        pyforms.Textbox("manager_code", rules.len_of(1, 32), description=u"客户经理工号",required="required",**input_style),
        pyforms.Textbox("manager_name", rules.len_of(1, 64), description=u"客户经理姓名",required="required",**input_style),
        pyforms.Textbox("manager_mobile",rules.len_of(1, 32), description=u"客户经理电话",required="required",**input_style),
        pyforms.Textbox("manager_email",rules.len_of(1, 255), description=u"客户经理邮箱",**input_style),
        pyforms.Dropdown("operator_name", description=u"关联操作员",args=oprs,**input_style),
        pyforms.Button("submit", type="submit", html=u"<b> 提交 </b>", **button_style),
        title=u"创建客户经理资料"
    )()

def cmanage_update_form(oprs=[]):
    return pyforms.Form(
        pyforms.Hidden("id",description=u"客户经理id"),
        pyforms.Textbox("manager_code", rules.len_of(1, 32), description=u"客户经理工号",readonly="readonly",**input_style),
        pyforms.Textbox("manager_name", rules.len_of(1, 64), description=u"客户经理姓名",required="required",**input_style),
        pyforms.Textbox("manager_mobile",rules.len_of(1, 32), description=u"客户经理电话",required="required",**input_style),
        pyforms.Textbox("manager_email",rules.len_of(1, 255), description=u"客户经理邮箱",**input_style),
        pyforms.Dropdown("operator_name", description=u"关联操作员",args=oprs,**input_style),
        pyforms.Button("submit", type="submit", html=u"<b> 提交 </b>", **button_style),
        title=u"更新客户经理资料"
    )()