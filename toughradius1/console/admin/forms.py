#coding:utf-8
from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style,input_style

boolean = {0:u"否", 1:u"是"}

booleans = {'0': u"否", '1': u"是"}
bool_bypass = {'0': u"免密码认证", '1': u"强制密码认证"}

param_form = pyforms.Form(
    pyforms.Textbox("system_name", description=u"管理系统名称", **input_style),
    pyforms.Textbox("customer_system_name", description=u"自助服务系统名称", **input_style),
    pyforms.Textbox("customer_system_url", description=u"自助服务系统网站地址", **input_style),
    pyforms.Dropdown("online_support", args=booleans.items(), description=u"开启在线支持功能",**input_style),
    pyforms.Dropdown("customer_must_active",args=booleans.items(), description=u"激活邮箱才能自助开户充值",hr=True, **input_style),
    pyforms.Textbox("weixin_qrcode", description=u"微信公众号二维码图片(宽度230px)", **input_style),
    pyforms.Textbox("service_phone", description=u"客户服务电话", **input_style),
    pyforms.Textbox("service_qq", description=u"客户服务QQ号码", **input_style),
    pyforms.Textbox("rcard_order_url", description=u"充值卡订购网站地址",hr=True,**input_style),
    pyforms.Textbox("expire_notify_days",rules.is_number, description=u"到期提醒提前天数", **input_style),
    pyforms.Textarea("expire_notify_tpl", description=u"到期提醒邮件模板",rows=3, **input_style),
    pyforms.Textbox("expire_notify_url", description=u"到期通知触发URL", **input_style),
    pyforms.Textbox("expire_session_timeout", description=u"到期用户下发最大会话时长(秒)", **input_style),
    pyforms.Textbox("expire_addrpool", description=u"到期提醒下发地址池", hr=True,**input_style),    
    pyforms.Textbox("smtp_server", description=u"SMTP服务器", **input_style),
    pyforms.Textbox("smtp_user", description=u"SMTP用户名", **input_style),
    pyforms.Textbox("smtp_pwd", description=u"SMTP密码",hr=True, **input_style),
    # pyforms.Textbox("smtp_sender", description=u"smtp发送人名称", **input_style),
    pyforms.Dropdown("is_debug",args=booleans.items(), description=u"开启DEBUG",**input_style),
    pyforms.Dropdown("radiusd_bypass",args=bool_bypass.items(), description=u"Radius认证模式", **input_style),
    pyforms.Dropdown("allow_show_pwd", args=booleans.items(),description=u"是否允许查询用户密码", **input_style),
    pyforms.Textbox("radiusd_address", description=u"Radius服务IP地址",**input_style),
    pyforms.Textbox("radiusd_admin_port",rules.is_number, description=u"Radius服务管理端口",**input_style),
    pyforms.Textbox("acct_interim_intelval",rules.is_number, description=u"Radius记账间隔(秒)",**input_style),
    pyforms.Textbox("max_session_timeout",rules.is_number, description=u"Radius最大会话时长(秒)", **input_style),
    pyforms.Textbox("reject_delay",rules.is_number, description=u"拒绝延迟时间(秒)(0-9)",**input_style),
    # pyforms.Textbox("portal_secret", description=u"portal登陆密钥", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/param/update"
)


passwd_update_form = pyforms.Form(
    pyforms.Textbox("operator_name", description=u"管理员名", size=32, readonly="readonly", **input_style),
    pyforms.Password("operator_pass", rules.len_of(6, 32), description=u"管理员新密码", size=32,value="", required="required", **input_style),
    pyforms.Password("operator_pass_chk", rules.len_of(6, 32), description=u"确认管理员新密码", size=32,value="", required="required", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"管理密码更新",
    action="/passwd"
)



    













