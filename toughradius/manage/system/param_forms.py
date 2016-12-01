#coding:utf-8
from toughradius.common import btforms
from toughradius.common.btforms import rules
from toughradius.common.btforms.rules import button_style,input_style
button_style ={"class":"btn btn-md bg-navy"}
boolean = {0:u"否", 1:u"是"}
booleans = {'0': u"否", '1': u"是"}
mailmodes = {'toughcloud': u"硬派云邮件服务", 'smtp': u"SMTP服务"}
bool_bypass = {'0': u"免密码认证", '1': u"强制密码认证"}
button_style_link = {"class": "btn btn-sm btn-link"}


sys_form = btforms.Form(
    btforms.Textbox("system_name", description=u"管理系统名称",help=u"管理系统名称,可以根据你的实际情况进行定制", **input_style),
    btforms.Textbox("system_ticket_expire_days", description=u"上网日志保留天数", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/admin/param/update?active=syscfg"
)

toughcloud_form = btforms.Form(
    btforms.Textarea("toughcloud_license", description=u"硬派云服务授权码", 
        help=u"硬派云服务授权码是用来接入硬派云的凭证，请妥善保管，泄露授权码会给您带来安全隐患。如果授权码泄露，请尽快申请新的授权码。",
        rows=3,**input_style),
    btforms.Textbox("toughcloud_service_mail", description=u"服务联系邮件", 
        help=u"如果启用硬派云邮件服务，该邮件可以作为发送地址和用户回复地址",
        **input_style),
    btforms.Textbox("toughcloud_service_call", description=u"服务联系电话", 
        help=u"如果启用硬派云邮件服务，该电话会出现在通知用户的邮件内容里",
        **input_style),
    btforms.Button("fetch_toughcloud_service", type="button", 
        html=u'<a href="https://www.toughcloud.net" target="_blank">申请硬派云服务</a>',**button_style_link),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/admin/param/update?active=toughcloudcfg"
)

notify_form = btforms.Form(
    btforms.Dropdown("mail_notify_enable", args=booleans.items(), description=u"启动邮件提醒", **input_style),
    btforms.Dropdown("sms_notify_enable", args=booleans.items(), description=u"启动短信提醒", **input_style),
    btforms.Dropdown("webhook_notify_enable", args=booleans.items(), description=u"启动URL触发到期提醒", **input_style),
    btforms.Textbox("expire_notify_days", rules.is_number, description=u"到期提醒提前天数", **input_style),
    btforms.Textbox("expire_notify_interval", rules.is_number, description=u"到期提醒间隔(分钟)", **input_style),
    btforms.Textbox("expire_notify_time", rules.is_time_hm, description=u"到期提醒时间(hh:mm)",
        help=u"优先于到期提醒间隔，格式：(几点:几分)", **input_style),
    btforms.Textbox("expire_notify_url", description=u"到期通知触发URL", **input_style),
    btforms.Textbox("expire_session_timeout", description=u"到期用户下发最大会话时长(秒)", **input_style),
    btforms.Textbox("expire_addrpool", description=u"到期提醒下发地址池", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/admin/param/update?active=notifycfg"
)



mail_form = btforms.Form(
    btforms.Dropdown("mail_mode", args=mailmodes.items(), description=u"邮件通知服务类型", **input_style),
    btforms.Textbox("smtp_server", description=u"SMTP服务器", **input_style),
    btforms.Textbox("smtp_port", description=u"SMTP服务器端口", **input_style),
    btforms.Textbox("smtp_from", description=u"SMTP邮件发送地址", **input_style),
    btforms.Textbox("smtp_sender", description=u"SMTP邮件发送人名称", **input_style),
    btforms.Textarea("smtp_notify_tpl", description=u"到期提醒模板", rows=5, **input_style),
    btforms.Textbox("smtp_user", description=u"SMTP用户名", **input_style),
    btforms.Textbox("smtp_pwd", description=u"SMTP密码", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"邮件设置",
    action="/admin/param/update?active=mailcfg"
)

rad_form = btforms.Form(
    btforms.Dropdown("radius_bypass", args=bool_bypass.items(), description=u"Radius认证模式", **input_style),
    btforms.Textbox("radius_acct_interim_intelval", rules.is_number, description=u"Radius记账间隔(秒)",help=u"radius向bas设备下发的全局记账间隔，bas不支持则不生效", **input_style),
    btforms.Textbox("radius_max_session_timeout", rules.is_number, description=u"Radius最大会话时长(秒)",help=u"用户在线达到最大会话时长时会自动断开", **input_style),
    btforms.Dropdown("radius_auth_auto_unlock", args=booleans.items(), description=u"并发自动解锁", help=u"如果账号被挂死，认证时自动踢下线",**input_style),
    btforms.Dropdown("radius_user_trace", args=booleans.items(), description=u"开启用户消息跟踪", help=u"开启此项会记录用户最近认证消息，可用于跟踪用户故障，参考用户管理-用户账号诊断",**input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"参数配置管理",
    action="/admin/param/update?active=radcfg"
)










