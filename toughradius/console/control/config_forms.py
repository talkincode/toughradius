# coding:utf-8
from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}
booleans = {'0': u"否", '1': u"是"}
timezones = {'CST-8':u"Asia/Shanghai"}

default_form = pyforms.Form(
    pyforms.Dropdown("debug", args=booleans.items(), description=u"开启DEBUG", help=u"开启此项，可以获取更多的系统日志纪录", **input_style),
    pyforms.Dropdown("tz", args=timezones.items(), description=u"时区", **input_style),
    pyforms.Textbox("secret", description=u"安全密钥", readonly="readonly", **input_style),
    pyforms.Dropdown("ssl", args=booleans.items(), description=u"开启SSL", help=u"开启此项，可以使用安全HTTP访问", **input_style),
    pyforms.Textbox("privatekey", description=u"安全证书路径", **input_style),
    pyforms.Textbox("certificate", description=u"安全证书签名路径", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"系统配置管理",
    action="/config/default/update"
)

dbtypes = {'mysql': u"mysql",'sqlite':u"sqlite"}

database_form = pyforms.Form(
    pyforms.Dropdown("echo", args=booleans.items(), description=u"开启数据库DEBUG", help=u"开启此项，可以在控制台打印SQL语句", **input_style),
    pyforms.Dropdown("dbtype", args=dbtypes.items(), description=u"数据库类型", **input_style),
    pyforms.Textbox("dburl", description=u"数据库连接字符串", **input_style),
    pyforms.Textbox("pool_size", description=u"连接池大小", **input_style),
    pyforms.Textbox("pool_recycle", description=u"连接池回收间隔（秒）", **input_style),
    pyforms.Textbox("backup_path", description=u"数据库备份路径", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"数据库配置管理",
    action="/config/database/update"
)

radiusd_form = pyforms.Form(
    pyforms.Textbox("host", description=u"radius认证计费监听地址", **input_style),
    pyforms.Textbox("authport", description=u"认证端口", **input_style),
    pyforms.Textbox("acctport", description=u"记账端口", **input_style),
    pyforms.Textbox("adminport", description=u"管理端口", **input_style),
    pyforms.Textbox("cache_timeout", description=u"缓存时间（秒）", **input_style),
    pyforms.Textbox("logfile", description=u"日志文件", readonly="readonly",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"radiusd配置管理",
    action="/config/radiusd/update"
)

admin_form = pyforms.Form(
    pyforms.Textbox("host", description=u"radius营业管理监听地址", **input_style),
    pyforms.Textbox("port", description=u"营业管理监听端口", **input_style),
    pyforms.Textbox("logfile", description=u"日志文件", readonly="readonly", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"radiusd营业管理配置管理",
    action="/config/admin/update"
)

customer_form = pyforms.Form(
    pyforms.Textbox("host", description=u"radius自助服务监听地址", **input_style),
    pyforms.Textbox("port", description=u"自助服务监听端口", **input_style),
    pyforms.Textbox("logfile", description=u"日志文件", readonly="readonly", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"radiusd自助服务配置管理",
    action="/config/customer/update"
)

control_form = pyforms.Form(
    pyforms.Textbox("host", description=u"radius系统控制监听地址", readonly="readonly", **input_style),
    pyforms.Textbox("port", description=u"系统控制监听端口", **input_style),
    pyforms.Textbox("logfile", description=u"日志文件", readonly="readonly", **input_style),
    pyforms.Textbox("user", description=u"管理员名", **input_style),
    pyforms.Password("passwd", description=u"管理密码", autocomplete="off",help=u"留空则不修改", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"radiusd系统控制配置管理",
    action="/config/control/update"
)
