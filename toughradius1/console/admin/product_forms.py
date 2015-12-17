#!/usr/bin/env python
#coding=utf-8

from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}

product_policy = {0: u'预付费包月', 1: u"预付费时长", 2: u"买断包月", 3: u"买断时长", 4: u"预付费流量", 5: u"买断流量"}
product_status_dict = {0: u'正常', 1: u"停用"}

def product_add_form():
    return pyforms.Form(
        pyforms.Textbox("product_name", rules.len_of(4, 64), description=u"资费名称", required="required", **input_style),
        pyforms.Dropdown("product_policy", args=product_policy.items(), description=u"计费策略", required="required",**input_style),
        pyforms.Textbox("fee_months", rules.is_number, description=u"买断授权月数", value=0, **input_style),
        pyforms.Textbox("fee_times", rules.is_number3, description=u"买断时长(小时)", value=0, **input_style),
        pyforms.Textbox("fee_flows", rules.is_number3, description=u"买断流量(MB)", value=0, **input_style),
        pyforms.Textbox("fee_price", rules.is_rmb, description=u"资费价格(元)", required="required", **input_style),
        pyforms.Hidden("fee_period", rules.is_period, description=u"开放认证时段", **input_style),
        pyforms.Textbox("concur_number", rules.is_numberOboveZore, description=u"并发数控制(0表示不限制)", value="0",**input_style),
        pyforms.Dropdown("bind_mac", args=boolean.items(), description=u"是否绑定MAC ", **input_style),
        pyforms.Dropdown("bind_vlan", args=boolean.items(), description=u"是否绑定VLAN ", **input_style),
        pyforms.Textbox("input_max_limit", rules.is_number3, description=u"最大上行速率(Mbps)", required="required",**input_style),
        pyforms.Textbox("output_max_limit", rules.is_number3, description=u"最大下行速率(Mbps)", required="required",**input_style),
        pyforms.Dropdown("product_status", args=product_status_dict.items(), description=u"资费状态", required="required",**input_style),
        pyforms.Button("submit", type="submit", id="submit", html=u"<b>提交</b>", **button_style),
        title=u"增加资费",
        action="/product/add"
    )


def product_update_form():
    return pyforms.Form(
        pyforms.Hidden("id", description=u"编号"),
        pyforms.Hidden("product_policy", description=u""),
        pyforms.Textbox("product_name", rules.len_of(4, 32), description=u"资费名称", required="required", **input_style),
        pyforms.Textbox("product_policy_name", description=u"资费策略", readonly="readonly", required="required",**input_style),
        pyforms.Dropdown("product_status", args=product_status_dict.items(), description=u"资费状态", required="required",**input_style),
        pyforms.Textbox("fee_months", rules.is_number, description=u"买断授权月数", value=0, **input_style),
        pyforms.Textbox("fee_times", rules.is_number3, description=u"买断时长(小时)", value=0, **input_style),
        pyforms.Textbox("fee_flows", rules.is_number3, description=u"买断流量(MB)", value=0, **input_style),
        pyforms.Textbox("fee_price", rules.is_rmb, description=u"资费价格(元)", required="required", **input_style),
        pyforms.Hidden("fee_period", rules.is_period, description=u"开放认证时段", **input_style),
        pyforms.Textbox("concur_number", rules.is_number, description=u"并发数控制(0表示不限制)", required="required",**input_style),
        pyforms.Dropdown("bind_mac", args=boolean.items(), description=u"是否绑定MAC", required="required", **input_style),
        pyforms.Dropdown("bind_vlan", args=boolean.items(), description=u"是否绑定VLAN", required="required",**input_style),
        pyforms.Textbox("input_max_limit", rules.is_number3, description=u"最大上行速率(Mbps)", required="required",**input_style),
        pyforms.Textbox("output_max_limit", rules.is_number3, description=u"最大下行速率(Mbps)", required="required",**input_style),
        pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
        title=u"修改资费",
        action="/product/update"
    )


product_attr_add_form = pyforms.Form(
    pyforms.Hidden("product_id", description=u"资费编号"),
    pyforms.Textbox("attr_name", rules.len_of(1, 255), description=u"策略名称", required="required", help=u"策略参考",**input_style),
    pyforms.Textbox("attr_value", rules.len_of(1, 255), description=u"策略值", required="required", **input_style),
    pyforms.Textbox("attr_desc", rules.len_of(1, 255), description=u"策略描述", required="required", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加策略属性",
    action="/product/attr/add"
)

product_attr_update_form = pyforms.Form(
    pyforms.Hidden("id", description=u"编号"),
    pyforms.Hidden("product_id", description=u"资费编号"),
    pyforms.Textbox("attr_name", rules.len_of(1, 255), description=u"策略名称", readonly="required", **input_style),
    pyforms.Textbox("attr_value", rules.len_of(1, 255), description=u"策略值", required="required", **input_style),
    pyforms.Textbox("attr_desc", rules.len_of(1, 255), description=u"策略描述", required="required", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改策略属性",
    action="/product/attr/update"
)

