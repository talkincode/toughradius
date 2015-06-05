#coding:utf-8
from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style,input_style

issues_types = {'0':u'新装','1':u'故障','2':u'投诉','3':u'其他'}

process_status = {'1': u'处理中', '2': u'挂起', '3': u'取消','4':u'处理完成'}

def issues_add_form(oprs=[]):
    return pyforms.Form(
        pyforms.Textbox("account_number", rules.len_of(1, 32), description=u"用户账号",required="required",**input_style),
        pyforms.Dropdown("issues_type", description=u"工单类型", args=issues_types.items(), **input_style),
        pyforms.Textarea("content", rules.len_of(1, 1024), description=u"工单内容", rows=6, required="required",**input_style),
        pyforms.Dropdown("assign_operator",  description=u"指派操作员",args=oprs, required="required", **input_style),
        pyforms.Button("submit", type="submit", html=u"<b> 提交 </b>", **button_style),
        action="/issues/add",
        title=u"创建用户工单"
    )()


def issues_process_form():
    return pyforms.Form(
        pyforms.Hidden("issues_id", rules.len_of(1, 32), description=u"工单id", required="required", **input_style),
        pyforms.Textarea("accept_result", rules.len_of(1, 1024), description=u"处理描述", rows=6, required="required",**input_style),
        pyforms.Dropdown("accept_status", description=u"处理结果", args=process_status.items(), required="required", **input_style),
        pyforms.Button("submit", type="submit", html=u"<b> 处理用户工单 </b>", **button_style),
        action="/issues/process",
        title=u"处理用户工单"
    )()

