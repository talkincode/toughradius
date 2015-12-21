#coding:utf-8
from toughradius.common import pyforms
from toughradius.common.pyforms import rules
from toughradius.common.pyforms.rules import button_style,input_style

boolean = {0:u"否", 1:u"是"}
booleans = {'0': u"否", '1': u"是"}
bool_bypass = {'0': u"免密码认证", '1': u"强制密码认证"}
ra_protocols = {'http': u"HTTP协议", 'zmq': u"ZMQ协议"}

sys_form = pyforms.Form(
    pyforms.Textbox("system_name", description=u"管理系统名称",help=u"管理系统名称,可以根据你的实际情况进行定制", **input_style),
    pyforms.Textbox("system_ticket_expire_days", description=u"上网日志保留天数", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/param/update?active=syscfg"
)

serv_form = pyforms.Form(
    pyforms.Textbox("customer_weixin_qrcode", description=u"微信公众号二维码图片(宽度230px)", **input_style),
    pyforms.Textbox("customer_service_phone", description=u"客户服务电话", **input_style),
    pyforms.Textbox("customer_service_qq", description=u"客户服务QQ号码", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/param/update?active=servcfg"
)

notify_form = pyforms.Form(
    pyforms.Textbox("expire_notify_days", rules.is_number, description=u"到期提醒提前天数", **input_style),
    pyforms.Textarea("expire_notify_tpl", description=u"到期提醒邮件模板", rows=5, **input_style),
    pyforms.Textbox("expire_notify_url", description=u"到期通知触发URL", **input_style),
    pyforms.Textbox("expire_session_timeout", description=u"到期用户下发最大会话时长(秒)", **input_style),
    pyforms.Textbox("expire_addrpool", description=u"到期提醒下发地址池", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/param/update?active=notifycfg"
)

mail_form = pyforms.Form(
    pyforms.Textbox("smtp_server", description=u"SMTP服务器", **input_style),
    pyforms.Textbox("smtp_port", description=u"SMTP服务器", **input_style),
    pyforms.Textbox("smtp_from", description=u"SMTP邮件发送地址", **input_style),
    pyforms.Textbox("smtp_user", description=u"SMTP用户名", **input_style),
    pyforms.Textbox("smtp_pwd", description=u"SMTP密码", help=u"如果密码不是必须的，请填写none", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/param/update?active=mailcfg"
)

rad_form = pyforms.Form(
    pyforms.Dropdown("radius_bypass", args=bool_bypass.items(), description=u"Radius认证模式", **input_style),
    pyforms.Dropdown("radius_agent_protocol", args=ra_protocols.items(), description=u"Radius认证接口协议", **input_style),
    pyforms.Textbox("radius_acct_interim_intelval", rules.is_number, description=u"Radius记账间隔(秒)",help=u"radius向bas设备下发的全局记账间隔，bas不支持则不生效", **input_style),
    pyforms.Textbox("radius_max_session_timeout", rules.is_number, description=u"Radius最大会话时长(秒)",help=u"用户在线达到最大会话时长时会自动断开", **input_style),
    pyforms.Textbox("radius_reject_delay", rules.is_number, description=u"拒绝延迟时间(秒)(0-9)",help=u"延迟拒绝消息的下发间隔，防御ddos攻击", **input_style),
    pyforms.Dropdown("radius_auth_auto_unlock", args=booleans.items(), description=u"并发自动解锁", help=u"如果账号被挂死，认证时自动踢下线",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/param/update?active=radcfg"
)










