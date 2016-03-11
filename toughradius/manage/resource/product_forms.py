#!/usr/bin/env python
#coding=utf-8

from toughlib import btforms
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib.btforms.rules import button_style, input_style
from toughradius.manage.settings import * 

boolean = {0: u"否", 1: u"是"}

product_policy = {
    PPMonth: u'预付费包月', 
    PPTimes: u"预付费时长", 
    BOMonth: u"买断包月", 
    BOTimes: u"买断时长", 
    PPFlow: u"预付费流量", 
    BOFlows: u"买断流量",
    FreeFee: u"自由资费"
}
product_status_dict = {0: u'正常', 1: u"停用"}

def product_add_form():
    return btforms.Form(
        btforms.Textbox("product_name", rules.len_of(4, 64), description=u"资费名称", required="required", **input_style),
        btforms.Dropdown("product_policy", args=product_policy.items(), description=u"计费策略", required="required",**input_style),
        btforms.Textbox("fee_months", rules.is_number, description=u"买断授权月数", value=0, **input_style),
        btforms.Textbox("fee_times", rules.is_number3, description=u"买断时长(小时)", value=0, **input_style),
        btforms.Textbox("fee_flows", rules.is_number3, description=u"买断流量(MB)", value=0, **input_style),
        btforms.Textbox("fee_price", rules.is_rmb, description=u"资费价格(元)", required="required", **input_style),
        # btforms.Hidden("fee_period", rules.is_period, description=u"开放认证时段", **input_style),
        btforms.Textbox("concur_number", rules.is_numberOboveZore, description=u"并发数控制(0表示不限制)", value="0",**input_style),
        btforms.Dropdown("bind_mac", args=boolean.items(), description=u"是否绑定MAC ", **input_style),
        btforms.Dropdown("bind_vlan", args=boolean.items(), description=u"是否绑定VLAN ", **input_style),
        btforms.Textbox("input_max_limit", rules.is_number3, description=u"最大上行速率(Mbps)", required="required",**input_style),
        btforms.Textbox("output_max_limit", rules.is_number3, description=u"最大下行速率(Mbps)", required="required",**input_style),
        btforms.Dropdown("product_status", args=product_status_dict.items(), description=u"资费状态", required="required",**input_style),
        btforms.Button("submit", type="submit", id="submit", html=u"<b>提交</b>", **button_style),
        title=u"增加资费",
        action="/admin/product/add"
    )


def product_update_form():
    return btforms.Form(
        btforms.Hidden("id", description=u"编号"),
        btforms.Hidden("product_policy", description=u""),
        btforms.Textbox("product_name", rules.len_of(4, 32), description=u"资费名称", required="required", **input_style),
        btforms.Textbox("product_policy_name", description=u"资费策略", readonly="readonly", required="required",**input_style),
        btforms.Dropdown("product_status", args=product_status_dict.items(), description=u"资费状态", required="required",**input_style),
        btforms.Textbox("fee_months", rules.is_number, description=u"买断授权月数", value=0, **input_style),
        btforms.Textbox("fee_times", rules.is_number3, description=u"买断时长(小时)", value=0, **input_style),
        btforms.Textbox("fee_flows", rules.is_number3, description=u"买断流量(MB)", value=0, **input_style),
        btforms.Textbox("fee_price", rules.is_rmb, description=u"资费价格(元)", required="required", **input_style),
        # btforms.Hidden("fee_period", rules.is_period, description=u"开放认证时段", **input_style),
        btforms.Textbox("concur_number", rules.is_number, description=u"并发数控制(0表示不限制)", required="required",**input_style),
        btforms.Dropdown("bind_mac", args=boolean.items(), description=u"是否绑定MAC", required="required", **input_style),
        btforms.Dropdown("bind_vlan", args=boolean.items(), description=u"是否绑定VLAN", required="required",**input_style),
        btforms.Textbox("input_max_limit", rules.is_number3, description=u"最大上行速率(Mbps)", required="required",**input_style),
        btforms.Textbox("output_max_limit", rules.is_number3, description=u"最大下行速率(Mbps)", required="required",**input_style),
        btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
        title=u"修改资费",
        action="/admin/product/update"
    )


product_attr_add_form = btforms.Form(
    btforms.Hidden("product_id", description=u"资费编号"),
    btforms.Textbox("attr_name", rules.len_of(1, 255), description=u"策略名称", required="required", help=u"策略参考",**input_style),
    btforms.Textbox("attr_value", rules.len_of(1, 255), description=u"策略值", required="required", **input_style),
    btforms.Textbox("attr_desc", rules.len_of(1, 255), description=u"策略描述", required="required", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加策略属性",
    action="/admin/product/attr/add"
)

product_attr_update_form = btforms.Form(
    btforms.Hidden("id", description=u"编号"),
    btforms.Hidden("product_id", description=u"资费编号"),
    btforms.Textbox("attr_name", rules.len_of(1, 255), description=u"策略名称", readonly="required", **input_style),
    btforms.Textbox("attr_value", rules.len_of(1, 255), description=u"策略值", required="required", **input_style),
    btforms.Textbox("attr_desc", rules.len_of(1, 255), description=u"策略描述", required="required", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改策略属性",
    action="/admin/product/attr/update"
)

