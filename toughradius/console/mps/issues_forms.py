#!/usr/bin/env python
#coding=utf-8
from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style,input_style


issues_types = [('0',u'办理业务'),('1',u"申报故障"),('2',u"投诉"),('3',u"其他")]

#工单申报表单
issues_add_form = pyforms.Form(
    pyforms.Hidden("openid", description=u"客户openid",  required="required", **input_style),
    pyforms.Dropdown("issues_type", description=u"工单类型", args=issues_types, **input_style),
    pyforms.Textarea("content", description=u"内容描述", rows="3", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提 交 工 单</b>",**button_style),
    title=u"工单申请"
)
