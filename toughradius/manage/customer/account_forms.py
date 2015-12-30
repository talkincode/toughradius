#!/usr/bin/env python
#coding=utf-8

from toughlib import btforms
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib.btforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}


def account_open_form(products=[]):
    return btforms.Form(
        btforms.Hidden("node_id", description=u"区域", **input_style),
        btforms.Hidden("customer_id", description=u"编号"),
        btforms.Textbox("realname", description=u"用户姓名", readonly="readonly", **input_style),
        btforms.Textbox("account_number", description=u"用户账号", required="required", **input_style),
        btforms.Textbox("password", description=u"认证密码", required="required", **input_style),
        btforms.Textbox("ip_address", description=u"用户IP地址", **input_style),
        btforms.Textbox("address", description=u"用户装机地址", **input_style),
        btforms.Dropdown("product_id", args=products, description=u"资费", required="required", **input_style),
        btforms.Textbox("months", rules.is_number, description=u"月数(包月有效)", required="required", **input_style),
        btforms.Textbox("fee_value", rules.is_rmb, description=u"缴费金额", required="required", **input_style),
        btforms.Textbox("expire_date", rules.is_date, description=u"过期日期", required="required", **input_style),
        btforms.Hidden("status", value=1, description=u"用户状态", **input_style),
        btforms.Textarea("account_desc", description=u"用户描述", rows=4, **input_style),
        btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户新开账号",
        action="/admin/account/open"
    )


def account_update_form():
    return btforms.Form(
        btforms.Textbox("account_number", description=u"用户账号", readonly="readonly", **input_style),
        btforms.Textbox("ip_address", description=u"用户IP地址", **input_style),
        btforms.Hidden("install_address", description=u"用户安装地址", **input_style),
        btforms.Textbox("new_password", description=u"认证密码(留空不修改)", **input_style),
        btforms.Textbox("user_concur_number", rules.is_number, description=u"用户并发数", required="required", **input_style),
        btforms.Dropdown("bind_mac", args=boolean.items(), description=u"是否绑定MAC", **input_style),
        btforms.Dropdown("bind_vlan", args=boolean.items(), description=u"是否绑定VLAN", **input_style),
        btforms.Textarea("account_desc", description=u"用户描述", rows=4, **input_style),
        btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户策略修改",
        action="/admin/account/update"
    )


account_next_form = btforms.Form(
        btforms.Hidden("product_id", description=u"资费"),
        btforms.Hidden("old_expire", description=u""),
        btforms.Hidden("account_number", description=u"用户账号"),
        btforms.Textbox("months", rules.is_number, description=u"月数(包月有效)", value=0, required="required",
                        **input_style),
        btforms.Textbox("fee_value", rules.is_rmb, description=u"缴费金额", value=0, required="required", **input_style),
        btforms.Textbox("expire_date", rules.is_date, description=u"过期日期", required="required", **input_style),
        btforms.Textarea("operate_desc", rules.len_of(0,512), description=u"操作描述", rows=4, **input_style),
        btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户续费",
        action="/admin/account/next"
    )


account_charge_form = btforms.Form(
    btforms.Hidden("account_number", description=u"用户账号", required="required", **input_style),
    btforms.Textbox("fee_value", rules.is_rmb, description=u"缴费金额", value=0, required="required", **input_style),
    btforms.Textarea("operate_desc", rules.len_of(0, 512),description=u"操作描述", rows=4, **input_style),
    btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"用户充值",
    action="/admin/account/charge"
)

account_cancel_form = btforms.Form(
    btforms.Hidden("account_number", description=u"用户账号", required="required", **input_style),
    btforms.Textbox("fee_value", rules.is_rmb, description=u"退费金额", required="required", **input_style),
    btforms.Textarea("operate_desc", rules.len_of(0, 512),description=u"操作描述", rows=4, **input_style),
    btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"用户销户",
    action="/admin/account/cancel"
)


def account_change_form(products=[]):
    return btforms.Form(
        btforms.Hidden("account_number", description=u"用户账号", required="required", **input_style),
        btforms.Dropdown("product_id", args=products, description=u"资费", required="required", **input_style),
        btforms.Textbox("add_value", rules.is_rmb, description=u"缴费金额", required="required", value="0", **input_style),
        btforms.Textbox("back_value", rules.is_rmb, description=u"退费金额", required="required", value="0", **input_style),
        btforms.Textbox("expire_date", rules.is_date, description=u"过期日期", value="0000-00-00", **input_style),
        btforms.Textbox("balance", rules.is_rmb, description=u"用户变更后余额", value="0.00", **input_style),
        btforms.Textbox("time_length", description=u"用户时长(小时)", value="0.00", **input_style),
        btforms.Textbox("flow_length", description=u"用户流量(MB)", value="0", **input_style),
        btforms.Textarea("operate_desc", rules.len_of(0, 512),description=u"操作描述", rows=4, **input_style),
        btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户资费变更",
        action="/admin/account/change"
    )