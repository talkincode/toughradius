#!/usr/bin/env python
# coding=utf-8

from toughlib import btforms
from toughlib.btforms import rules
from toughlib.btforms.rules import button_style, input_style


password_update_form = btforms.Form(
    btforms.Textbox("tr_user", description=u"管理员名", size=32, readonly="readonly", **input_style),
    btforms.Password("tr_user_pass", rules.len_of(6, 32), description=u"管理员新密码", size=32,value="", required="required", **input_style),
    btforms.Password("tr_user_pass_chk", rules.len_of(6, 32), description=u"确认管理员新密码", size=32,value="", required="required", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"管理密码更新",
    action="/admin/password"
)
