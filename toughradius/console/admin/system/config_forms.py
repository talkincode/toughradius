# coding:utf-8
from toughradius.common import pyforms
from toughradius.common.pyforms import dataform
from toughradius.common.pyforms import rules
from toughradius.common.pyforms.rules import button_style, input_style

boolean = {0: u"否", 1: u"是"}
booleans = {'0': u"否", '1': u"是"}
timezones = {'CST-8':u"Asia/Shanghai"}

default_form = pyforms.Form(
    pyforms.Dropdown("debug", args=booleans.items(), description=u"开启DEBUG", help=u"开启此项，可以获取更多的系统日志纪录", **input_style),
    pyforms.Dropdown("tz", args=timezones.items(), description=u"时区", **input_style),
    pyforms.Textbox("secret", description=u"安全密钥", readonly="readonly", **input_style),
    # pyforms.Dropdown("ssl", args=booleans.items(), description=u"开启SSL", help=u"开启此项，可以使用安全HTTP访问", **input_style),
    # pyforms.Textbox("privatekey", description=u"安全证书路径", **input_style),
    # pyforms.Textbox("certificate", description=u"安全证书签名路径", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"系统配置管理",
    action="/admin/config/default/update"
)

dbtypes = {'mysql': u"mysql",'sqlite':u"sqlite"}

database_form = pyforms.Form(
    pyforms.Dropdown("echo", args=booleans.items(), description=u"开启数据库DEBUG", help=u"开启此项，可以在控制台打印SQL语句", **input_style),
    pyforms.Textbox("dbtype",description=u"数据库类型", readonly="readonly",**input_style),
    pyforms.Textbox("dburl", description=u"数据库连接字符串", readonly="readonly", **input_style),
    pyforms.Textbox("pool_size", description=u"连接池大小", **input_style),
    pyforms.Textbox("pool_recycle", description=u"连接池回收间隔（秒）", **input_style),
    pyforms.Textbox("backup_path", description=u"数据库备份路径", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"数据库配置管理",
    action="/admin/config/database/update"
)

