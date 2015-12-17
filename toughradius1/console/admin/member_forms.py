#!/usr/bin/env python
#coding=utf-8

from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}
user_state = {1: u"正常", 2: u"停机", 3: u"销户", 4: u"到期"}
bind_state = {0: u"不绑定", 1: u"绑定"}


def user_open_form(nodes=[], products=[]):
    return pyforms.Form(
        pyforms.Dropdown("node_id", description=u"区域", args=nodes, required="required", **input_style),
        pyforms.Textbox("realname", rules.len_of(2, 32), description=u"用户姓名", required="required", **input_style),
        pyforms.Checkbox("is_samename", description=u"启用独立的自助服务用户名", checked=""),
        pyforms.Textbox("member_name", rules.len_of(0, 64), description=u"自助服务用户名", **input_style),
        pyforms.Textbox("member_password", rules.len_of(0, 128), description=u"自助服务用户密码", **input_style),
        pyforms.Textbox("idcard", rules.len_of(0, 32), description=u"证件号码", **input_style),
        pyforms.Textbox("mobile", rules.len_of(0, 32), description=u"用户手机号码", **input_style),
        pyforms.Textbox("address", description=u"用户地址", hr=True, **input_style),
        pyforms.Textbox("account_number", description=u"用户账号", required="required", **input_style),
        pyforms.Textbox("password", description=u"认证密码", required="required", **input_style),
        pyforms.Textbox("ip_address", description=u"用户IP地址", **input_style),
        pyforms.Dropdown("product_id", args=products, description=u"资费", required="required", **input_style),
        pyforms.Textbox("months", rules.is_number, description=u"月数(包月有效)", required="required", **input_style),
        pyforms.Textbox("fee_value", rules.is_rmb, description=u"缴费金额", required="required", **input_style),
        pyforms.Textbox("expire_date", rules.is_date, description=u"过期日期", required="required", **input_style),
        pyforms.Hidden("status", value=1, description=u"用户状态", **input_style),
        pyforms.Textarea("member_desc", description=u"用户描述", rows=4, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户开户",
        action="/member/open"
    )


def user_import_form(nodes=[], products=[]):
    return pyforms.Form(
        pyforms.Dropdown("node_id", description=u"用户区域", args=nodes, **input_style),
        pyforms.Dropdown("product_id", args=products, description=u"用户资费", required="required", **input_style),
        pyforms.File("import_file", description=u"用户数据文件", required="required", **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>立即导入</b>", **button_style),
        title=u"用户数据导入",
        action="/member/import"
    )


user_import_vform = dataform.Form(
    dataform.Item("realname", rules.not_null, description=u"用户姓名"),
    dataform.Item("idcard", rules.len_of(0, 32), description=u"证件号码"),
    dataform.Item("mobile", rules.len_of(0, 32), description=u"用户手机号码"),
    dataform.Item("address", description=u"用户地址"),
    dataform.Item("account_number", rules.not_null, description=u"用户账号"),
    dataform.Item("password", rules.not_null, description=u"用户密码"),
    dataform.Item("begin_date", rules.is_date, description=u"开通日期"),
    dataform.Item("expire_date", rules.is_date, description=u"过期日期"),
    dataform.Item("balance", rules.is_rmb, description=u"用户余额"),
    dataform.Item("time_length", description=u"用户时长"),
    dataform.Item("flow_length", description=u"用户流量"),
    title="import"
)


def member_update_form(nodes=[]):
    return pyforms.Form(
        pyforms.Hidden("account_number", description=u"用户账号"),
        pyforms.Hidden("member_id", description=u"mid", required="required", **input_style),
        pyforms.Textbox("realname", rules.len_of(2, 32), description=u"用户姓名", required="required", **input_style),
        pyforms.Textbox("member_name", description=u"自助服务用户名", readonly="readonly", autocomplete="off", **input_style),
        pyforms.Password("new_password", rules.len_of(0, 128), value="", description=u"自助服务密码(留空不修改)", **input_style),
        pyforms.Textbox("email", rules.len_of(0, 128), description=u"电子邮箱", **input_style),
        pyforms.Textbox("idcard", rules.len_of(0, 32), description=u"证件号码", **input_style),
        pyforms.Textbox("mobile", rules.len_of(0, 32), description=u"用户手机号码", **input_style),
        pyforms.Textbox("address", description=u"用户地址", hr=True, **input_style),
        pyforms.Textarea("member_desc", description=u"用户描述", rows=4, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户基本信息修改",
        action="/member/update"
    )