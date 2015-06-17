#!/usr/bin/env python
#coding=utf-8

from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}
card_types = {0: u'资费卡', 1: u'余额卡'}
card_states = {0: u'未激活', 1: u'已激活', 2: u"已使用", 3: u"已回收"}


def recharge_card_form(products=[]):
    return pyforms.Form(
        pyforms.Dropdown("card_type", args=card_types.items(), description=u"充值卡类型", **input_style),
        pyforms.Textbox("batch_no", rules.is_number, maxlength=8, description=u"批次号(年+月+2位序号，如：20150201)",required="required", **input_style),
        pyforms.Dropdown("product_id", args=products, description=u"资费", **input_style),
        pyforms.Textbox("start_no", rules.is_number, maxlength=5, description=u"开始卡号(最大5位)", required="required",**input_style),
        pyforms.Textbox("stop_no", rules.is_number, maxlength=5, description=u"结束卡号(最大5位)", required="required",**input_style),
        pyforms.Textbox("pwd_len", rules.is_number, description=u"密码长度(最大为16)", value=8, **input_style),
        pyforms.Textbox("fee_value", rules.is_rmb, description=u"面值/销售价(元)", value=0, **input_style),
        pyforms.Textbox("months", rules.is_number, description=u"授权时间(月)", readonly="readonly", value=0, **input_style),
        pyforms.Textbox("times", description=u"总时长(小时)", readonly="readonly", value=0, **input_style),
        pyforms.Textbox("flows", description=u"总流量(MB)", readonly="readonly", value=0, **input_style),
        pyforms.Textbox("expire_date", rules.is_date, description=u"过期时间", required="required", **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
        title=u"充值卡生成",
        action="/card/create"
    )
