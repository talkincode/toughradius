# coding:utf-8
from toughlib import btforms
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib.btforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}
booleans = {'0': u"否", '1': u"是"}
timezones = {'CST-8':u"Asia/Shanghai"}

loglevels = {
    'INFO': u"一般",
    'DEBUG': u"调试",
    'WARNING': u"警告",
    'ERROR': u"错误"
}

system_form = btforms.Form(
    btforms.Dropdown("debug", args=boolean.items(), description=u"开启DEBUG", help=u"开启此项，可以获取更多的系统日志纪录", **input_style),
    btforms.Dropdown("tz", args=timezones.items(), description=u"时区", **input_style),
    btforms.Textbox("secret", description=u"安全密钥", **input_style),
    # btforms.Button("regen_secret", type="button", onclick="update_secret();", html=u"<b>重新生成安全密钥</b>"),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"系统配置管理",
    action="/admin/config/system/update"
)

dbtypes = {'mysql': u"mysql",'sqlite':u"sqlite"}

database_form = btforms.Form(
    btforms.Dropdown("echo", args=boolean.items(), description=u"开启数据库DEBUG", help=u"开启此项，可以在控制台打印SQL语句", **input_style),
    btforms.Textbox("dbtype",description=u"数据库类型", readonly="readonly",**input_style),
    btforms.Textbox("dburl", description=u"数据库连接字符串", readonly="readonly", **input_style),
    btforms.Textbox("pool_size", description=u"连接池大小", **input_style),
    btforms.Textbox("pool_recycle", description=u"连接池回收间隔（秒）", **input_style),
    # btforms.Textbox("backup_path", description=u"数据库备份路径", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"数据库配置管理",
    action="/admin/config/database/update"
)

syslog_form = btforms.Form(
    btforms.Dropdown("enable", args=boolean.items(), description=u"开启syslog", **input_style),
    btforms.Textbox("server", description=u"syslog 服务器", **input_style),
    btforms.Textbox("port", description=u"syslog 服务端口(UDP)", **input_style),
    btforms.Dropdown("level", args=loglevels.items(), description=u"日志级别", **input_style),
    btforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"syslog 配置管理",
    action="/admin/config/syslog/update"
)

