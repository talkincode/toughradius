#!/usr/bin/env python
#coding=utf-8

from toughlib import btforms
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib.btforms.rules import button_style, input_style
from toughradius.manage.settings import * 

tpl_types = {
    OpenNotify : u'用户开户通知模板',
    NextNotify : u'用户续费通知模板',
    ExpireNotify : u'用户到期通知模板',
    # InstallNotify : u'装机工单通知模板',
    # MaintainNotify : u'维修工单通知模板',
    # OpenNoteRemark : u'开户票据打印备注模板',
    # NextNotePrint : u'续费票据打印备注模板',
    # RefundNodeRemark : u'退费票据打印备注模板'
}

content_tpl_add_form = btforms.Form(
    btforms.Dropdown("tpl_type", args=tpl_types.items(), description=u"模板类型", required="required", **input_style),
    btforms.Textarea("tpl_content", rules.len_of(2, 128), description=u"模板内容", rows=7,required="required", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加模板",
    action="/admin/contenttpl/add"
)

content_tpl_update_form = btforms.Form(
    btforms.Hidden("id", description=u"模板ID"),
    btforms.Dropdown("tpl_type", args=tpl_types.items(), description=u"模板类型", **input_style),
    btforms.Textarea("tpl_content", rules.len_of(2, 128), description=u"模板内容",rows=7, required="required", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改模板",
    action="/admin/contenttpl/update"
)
