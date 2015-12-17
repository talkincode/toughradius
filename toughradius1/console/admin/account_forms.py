#!/usr/bin/env python
#coding=utf-8

from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}


def account_open_form(products=[]):
    return pyforms.Form(
        pyforms.Hidden("node_id", description=u"区域", **input_style),
        pyforms.Hidden("member_id", description=u"编号"),
        pyforms.Textbox("realname", description=u"用户姓名", readonly="readonly", **input_style),
        pyforms.Textbox("account_number", description=u"用户账号", required="required", **input_style),
        pyforms.Textbox("password", description=u"认证密码", required="required", **input_style),
        pyforms.Textbox("ip_address", description=u"用户IP地址", **input_style),
        pyforms.Textbox("address", description=u"用户装机地址", **input_style),
        pyforms.Dropdown("product_id", args=products, description=u"资费", required="required", **input_style),
        pyforms.Textbox("months", rules.is_number, description=u"月数(包月有效)", required="required", **input_style),
        pyforms.Textbox("fee_value", rules.is_rmb, description=u"缴费金额", required="required", **input_style),
        pyforms.Textbox("expire_date", rules.is_date, description=u"过期日期", required="required", **input_style),
        pyforms.Hidden("status", value=1, description=u"用户状态", **input_style),
        pyforms.Textarea("account_desc", description=u"用户描述", rows=4, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户新开账号",
        action="/account/open"
    )


def account_update_form():
    return pyforms.Form(
        pyforms.Textbox("account_number", description=u"用户账号", readonly="readonly", **input_style),
        pyforms.Textbox("ip_address", description=u"用户IP地址", **input_style),
        pyforms.Hidden("install_address", description=u"用户安装地址", **input_style),
        pyforms.Textbox("new_password", description=u"认证密码(留空不修改)", **input_style),
        pyforms.Textbox("user_concur_number", rules.is_number, description=u"用户并发数", required="required",
                        **input_style),
        pyforms.Dropdown("bind_mac", args=boolean.items(), description=u"是否绑定MAC", **input_style),
        pyforms.Dropdown("bind_vlan", args=boolean.items(), description=u"是否绑定VLAN", **input_style),
        pyforms.Textarea("account_desc", description=u"用户描述", rows=4, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户变更资料",
        action="/account/update"
    )


account_next_form = pyforms.Form(
        pyforms.Hidden("product_id", description=u"资费"),
        pyforms.Hidden("old_expire", description=u""),
        pyforms.Hidden("account_number", description=u"用户账号"),
        pyforms.Textbox("months", rules.is_number, description=u"月数(包月有效)", value=0, required="required",
                        **input_style),
        pyforms.Textbox("fee_value", rules.is_rmb, description=u"缴费金额", value=0, required="required", **input_style),
        pyforms.Textbox("expire_date", rules.is_date, description=u"过期日期", required="required", **input_style),
        pyforms.Textarea("operate_desc", rules.len_of(0,512), description=u"操作描述", rows=4, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户续费",
        action="/account/next"
    )


account_charge_form = pyforms.Form(
    pyforms.Hidden("account_number", description=u"用户账号", required="required", **input_style),
    pyforms.Textbox("fee_value", rules.is_rmb, description=u"缴费金额", value=0, required="required", **input_style),
    pyforms.Textarea("operate_desc", rules.len_of(0, 512),description=u"操作描述", rows=4, **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"用户充值",
    action="/account/charge"
)

account_cancel_form = pyforms.Form(
    pyforms.Hidden("account_number", description=u"用户账号", required="required", **input_style),
    pyforms.Textbox("fee_value", rules.is_rmb, description=u"退费金额", required="required", **input_style),
    pyforms.Textarea("operate_desc", rules.len_of(0, 512),description=u"操作描述", rows=4, **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"用户销户",
    action="/account/cancel"
)


def account_change_form(products=[]):
    return pyforms.Form(
        pyforms.Hidden("account_number", description=u"用户账号", required="required", **input_style),
        pyforms.Dropdown("product_id", args=products, description=u"资费", required="required", **input_style),
        pyforms.Textbox("add_value", rules.is_rmb, description=u"缴费金额", required="required", value="0", **input_style),
        pyforms.Textbox("back_value", rules.is_rmb, description=u"退费金额", required="required", value="0", **input_style),
        pyforms.Textbox("expire_date", rules.is_date, description=u"过期日期", value="0000-00-00", **input_style),
        pyforms.Textbox("balance", rules.is_rmb, description=u"用户变更后余额", value="0.00", **input_style),
        pyforms.Textbox("time_length", description=u"用户时长(小时)", value="0.00", **input_style),
        pyforms.Textbox("flow_length", description=u"用户流量(MB)", value="0", **input_style),
        pyforms.Textarea("operate_desc", rules.len_of(0, 512),description=u"操作描述", rows=4, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户资费变更",
        action="/account/change"
    )