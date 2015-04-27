#coding:utf-8
from toughradius.console.libs import pyforms
from toughradius.console.libs.pyforms import dataform
from toughradius.console.libs.pyforms import rules
from toughradius.console.libs.pyforms.rules import button_style,input_style

boolean = {0:u"否", 1:u"是"}

param_form = pyforms.Form(
    pyforms.Textbox("system_name", description=u"管理系统名称", **input_style),
    pyforms.Textbox("customer_system_name", description=u"自助服务系统名称", **input_style),
    pyforms.Textbox("customer_system_url", description=u"自助服务系统网站地址", **input_style),
    pyforms.Textbox("customer_must_active", description=u"激活邮箱才能自助开户充值(0:否|1:是)",hr=True, **input_style),
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
    pyforms.Textbox("is_debug", description=u"DEBUG模式(0|1)",**input_style),
    pyforms.Textbox("radiusd_address", description=u"Radius服务IP地址",**input_style),
    pyforms.Textbox("radiusd_admin_port",rules.is_number, description=u"Radius服务管理端口",**input_style),
    pyforms.Textbox("acct_interim_intelval",rules.is_number, description=u"Radius记账间隔(秒)",**input_style),
    pyforms.Textbox("max_session_timeout",rules.is_number, description=u"Radius最大会话时长(秒)", **input_style),
    pyforms.Textbox("reject_delay",rules.is_number, description=u"拒绝延迟时间(秒)(0-9)",**input_style),
    pyforms.Textbox("portal_secret", description=u"portal登陆密钥", **input_style),
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

node_add_form = pyforms.Form(
    pyforms.Textbox("node_name", rules.len_of(2,32), description=u"区域名称",required="required",**input_style),
    pyforms.Textarea("node_desc", rules.len_of(0, 128), description=u"区域描述",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>",**button_style),
    title=u"增加区域",
    action="/node/add"
)

node_update_form = pyforms.Form(
    pyforms.Hidden("id",  description=u"区域编号"),
    pyforms.Textbox("node_name", rules.len_of(2, 32), description=u"区域名称", required="required",**input_style),
    pyforms.Textarea("node_desc", rules.len_of(0, 128), description=u"区域描述", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改区域",
    action="/node/update"
)

timetype = {0:u"标准时区,北京时间",1:u"时区和时间同区"}
bastype = {
            '0' : u'标准',
            '9' : u'思科',
            '3041' : u'阿尔卡特',
            '2352' : u'爱立信',
            '2011' : u'华为',
            '25506' : u'H3C',
            '3902' : u'中兴',
            '10055' : u'爱快',
            '14988' : u'RouterOS'
        }

bas_add_form = pyforms.Form(
        pyforms.Textbox("ip_addr", rules.is_ip, description=u"BAS地址",required="required",**input_style),
        pyforms.Textbox("bas_name", rules.len_of(2,64), description=u"BAS名称",required="required",**input_style),
        pyforms.Textbox("bas_secret", rules.is_alphanum2(4, 32), description=u"共享秘钥", required="required",**input_style),
        pyforms.Dropdown("vendor_id", description=u"BAS类型", args=bastype.items(), required="required",**input_style),
        pyforms.Textbox("coa_port", rules.is_number,description=u"CoA端口", default=3799,required="required",**input_style),
        pyforms.Dropdown("time_type", description=u"时间类型", args=timetype.items(), required="required",**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"增加BAS",
        action="/bas/add"
    )

bas_update_form = pyforms.Form(
        pyforms.Hidden("id",  description=u"编号"),
        pyforms.Textbox("ip_addr", rules.is_ip, description=u"BAS地址",  readonly="readonly",**input_style),
        pyforms.Textbox("bas_name", rules.len_of(2,64), description=u"BAS名称", required="required",**input_style),
        pyforms.Textbox("bas_secret", rules.is_alphanum2(4, 32), description=u"共享秘钥", required="required",**input_style),
        pyforms.Dropdown("vendor_id", description=u"BAS类型", args=bastype.items(), required="required",**input_style),
        pyforms.Textbox("coa_port", rules.is_number,description=u"CoA端口", default=3799,required="required",**input_style),
        pyforms.Dropdown("time_type", description=u"时间类型", args=timetype.items(), required="required",**input_style),
        pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
        title=u"修改BAS",
        action="/bas/update"
    )
    
opr_type = {0:u'系统管理员',1:u"普通操作员"}    
opr_status_dict = {0:u'正常',1:u"停用"}
    
def opr_add_form(nodes=[]):
    return pyforms.Form(
        pyforms.Textbox("operator_name", rules.len_of(2,32), description=u"操作员名称",required="required",**input_style),
        pyforms.Textbox("operator_desc", rules.len_of(0,255),description=u"操作员姓名",**input_style),
        pyforms.Password("operator_pass", rules.len_of(6, 128), description=u"操作员密码", required="required",**input_style),
        pyforms.Dropdown("operator_status", description=u"操作员状态", args=opr_status_dict.items(), required="required",**input_style),
        pyforms.Dropdown("operator_nodes", description=u"关联区域(多选)",args=nodes, required="required",multiple="multiple",size=4,**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"增加操作员",
        action="/opr/add"
    )  
    
def opr_update_form(nodes=[]):
    return pyforms.Form(
        pyforms.Hidden("id",  description=u"编号"),
        pyforms.Textbox("operator_name", description=u"操作员名称",readonly="readonly",**input_style),
        pyforms.Textbox("operator_desc", rules.len_of(0,255),description=u"操作员姓名",**input_style),
        pyforms.Password("operator_pass", rules.len_of(0, 128), description=u"操作员密码(留空不修改)",**input_style),
        pyforms.Dropdown("operator_status", description=u"操作员状态", args=opr_status_dict.items(), required="required",**input_style),
        pyforms.Dropdown("operator_nodes", description=u"关联区域(多选)",args=nodes, required="required",multiple="multiple",size=4,**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"修改操作员",
        action="/opr/update"
    )        

product_policy = {0:u'预付费包月',1:u"预付费时长",2:u"买断包月",3:u"买断时长",4:u"预付费流量",5:u"买断流量"}
product_status_dict = {0:u'正常',1:u"停用"}

def product_add_form():
    return pyforms.Form(
        pyforms.Textbox("product_name", rules.len_of(4, 64), description=u"资费名称",  required="required",**input_style),
        pyforms.Dropdown("product_policy", args=product_policy.items(), description=u"计费策略", required="required",**input_style),
        pyforms.Textbox("fee_months", rules.is_number, description=u"买断授权月数",value=0, **input_style),
        pyforms.Textbox("fee_times", rules.is_number, description=u"买断时长(小时)",value=0, **input_style),
        pyforms.Textbox("fee_flows", rules.is_number, description=u"买断流量(MB)",value=0, **input_style),
        pyforms.Textbox("fee_price", rules.is_rmb, description=u"资费价格(元)", required="required", **input_style),
        pyforms.Hidden("fee_period",rules.is_period,description=u"开放认证时段",**input_style),
        pyforms.Textbox("concur_number", rules.is_numberOboveZore,description=u"并发数控制(0表示不限制)",value="0", **input_style),
        pyforms.Dropdown("bind_mac",  args=boolean.items(), description=u"是否绑定MAC ",**input_style),
        pyforms.Dropdown("bind_vlan",  args=boolean.items(),description=u"是否绑定VLAN ",**input_style),
        pyforms.Textbox("input_max_limit",  rules.is_number3,description=u"最大上行速率(Mbps)",required="required", **input_style),
        pyforms.Textbox("output_max_limit",  rules.is_number3,description=u"最大下行速率(Mbps)",required="required",**input_style),
        pyforms.Dropdown("product_status", args=product_status_dict.items(),description=u"资费状态", required="required", **input_style),
        pyforms.Button("submit", type="submit",id="submit",html=u"<b>提交</b>", **button_style),
        title=u"增加资费",
        action="/product/add"
    )

def product_update_form():
    return pyforms.Form(
        pyforms.Hidden("id",  description=u"编号"),
        pyforms.Hidden("product_policy",  description=u""),
        pyforms.Textbox("product_name", rules.len_of(4, 32), description=u"资费名称", required="required",**input_style),
        pyforms.Textbox("product_policy_name", description=u"资费策略",readonly="readonly", required="required",**input_style),
        pyforms.Dropdown("product_status",args=product_status_dict.items(), description=u"资费状态", required="required", **input_style),
        pyforms.Textbox("fee_months", rules.is_number, description=u"买断授权月数",value=0, **input_style),
        pyforms.Textbox("fee_times", rules.is_number, description=u"买断时长(小时)",value=0, **input_style),
        pyforms.Textbox("fee_flows", rules.is_number, description=u"买断流量(MB)",value=0, **input_style),
        pyforms.Textbox("fee_price", rules.is_rmb,description=u"资费价格(元)", required="required", **input_style),
        pyforms.Hidden("fee_period", rules.is_period,description=u"开放认证时段",**input_style),
        pyforms.Textbox("concur_number", rules.is_number,description=u"并发数控制(0表示不限制)",required="required", **input_style),
        pyforms.Dropdown("bind_mac",  args=boolean.items(), description=u"是否绑定MAC",required="required",**input_style),
        pyforms.Dropdown("bind_vlan",  args=boolean.items(),description=u"是否绑定VLAN",required="required",**input_style),
        pyforms.Textbox("input_max_limit",  rules.is_number3,description=u"最大上行速率(Mbps)",required="required", **input_style),
        pyforms.Textbox("output_max_limit",  rules.is_number3,description=u"最大下行速率(Mbps)",required="required",**input_style),
        pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
        title=u"修改资费",
        action="/product/update"
    )

product_attr_add_form = pyforms.Form(
    pyforms.Hidden("product_id",  description=u"资费编号"),
    pyforms.Textbox("attr_name", rules.len_of(1, 255), description=u"策略名称",  required="required",help=u"策略参考",**input_style),
    pyforms.Textbox("attr_value", rules.len_of(1, 255),description=u"策略值", required="required",**input_style),
    pyforms.Textbox("attr_desc", rules.len_of(1, 255),description=u"策略描述", required="required",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加策略属性",
    action="/product/attr/add"
)

product_attr_update_form = pyforms.Form(
    pyforms.Hidden("id",  description=u"编号"),
    pyforms.Hidden("product_id",  description=u"资费编号"),
    pyforms.Textbox("attr_name", rules.len_of(1, 255), description=u"策略名称",  readonly="required",**input_style),
    pyforms.Textbox("attr_value", rules.len_of(1, 255),description=u"策略值", required="required",**input_style),
    pyforms.Textbox("attr_desc", rules.len_of(1, 255),description=u"策略描述", required="required",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改策略属性",
    action="/product/attr/update"
)


roster_type = {0:u"白名单", 1:u"黑名单"}

roster_add_form = pyforms.Form(
        pyforms.Textbox("mac_addr", description=u"MAC地址",required="required",**input_style),
        pyforms.Textbox("begin_time",description=u"开始时间", required="required",**input_style),
        pyforms.Textbox("end_time", description=u"结束时间", required="required",**input_style),
        pyforms.Dropdown("roster_type", args=roster_type.items(),description=u"类型",value=0, required="required",**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"增加黑白名单",
        action="/roster/add"
    )

roster_update_form = pyforms.Form(
    pyforms.Hidden("id",  description=u"编号"),
    pyforms.Textbox("mac_addr", description=u"MAC地址",readonly="readonly",**input_style),
    pyforms.Textbox("begin_time",description=u"开始时间", required="required",**input_style),
    pyforms.Textbox("end_time", description=u"结束时间", required="required",**input_style),
    pyforms.Dropdown("roster_type", args=roster_type.items(),description=u"类型",value=0, required="required",**input_style),
    pyforms.Button("submit",  type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改黑白名单",
    action="/roster/update"
)


user_state = {1:u"正常", 2:u"停机" , 3:u"销户", 4:u"到期"}
bind_state = {0: u"不绑定", 1: u"绑定"}

def user_open_form(nodes=[],products=[]):
    return pyforms.Form(
        pyforms.Dropdown("node_id", description=u"区域", args=nodes,required="required", **input_style),
        pyforms.Textbox("realname", rules.len_of(2,32), description=u"用户姓名", required="required",**input_style),
        pyforms.Checkbox("is_samename",description=u"启用独立的自助服务用户名", checked=""),
        pyforms.Textbox("member_name", rules.len_of(0,64), description=u"自助服务用户名", **input_style),
        pyforms.Textbox("member_password", rules.len_of(0,128), description=u"自助服务用户密码", **input_style),
        pyforms.Textbox("idcard", rules.len_of(0,32), description=u"证件号码", **input_style),
        pyforms.Textbox("mobile", rules.len_of(0,32),description=u"用户手机号码", **input_style),
        pyforms.Textbox("address", description=u"用户地址",hr=True, **input_style),
        pyforms.Textbox("account_number", description=u"用户账号",  required="required", **input_style),
        pyforms.Textbox("password", description=u"认证密码", required="required", **input_style),
        pyforms.Textbox("ip_address", description=u"用户IP地址",**input_style),
        pyforms.Dropdown("product_id",args=products, description=u"资费",  required="required", **input_style),
        pyforms.Textbox("months",rules.is_number, description=u"月数(包月有效)", required="required", **input_style),
        pyforms.Textbox("fee_value",rules.is_rmb, description=u"缴费金额",  required="required", **input_style),
        pyforms.Textbox("expire_date", rules.is_date,description=u"过期日期",  required="required", **input_style),
        pyforms.Hidden("status",value=1, description=u"用户状态",  **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户开户",
        action="/bus/member/open"
    )

def account_open_form(products=[]):
    return pyforms.Form(
        pyforms.Hidden("node_id", description=u"区域", **input_style),
        pyforms.Hidden("member_id",  description=u"编号"),
        pyforms.Textbox("realname", description=u"用户姓名", readonly="readonly",**input_style),
        pyforms.Textbox("account_number", description=u"用户账号",  required="required", **input_style),
        pyforms.Textbox("password", description=u"认证密码", required="required", **input_style),
        pyforms.Textbox("ip_address", description=u"用户IP地址",**input_style),
        pyforms.Textbox("address", description=u"用户装机地址",**input_style),
        pyforms.Dropdown("product_id",args=products, description=u"资费",  required="required", **input_style),
        pyforms.Textbox("months",rules.is_number, description=u"月数(包月有效)", required="required", **input_style),
        pyforms.Textbox("fee_value",rules.is_rmb, description=u"缴费金额",  required="required", **input_style),
        pyforms.Textbox("expire_date", rules.is_date,description=u"过期日期",  required="required", **input_style),
        pyforms.Hidden("status",value=1, description=u"用户状态",  **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户新开账号",
        action="/bus/account/open"
    )

def user_import_form(nodes=[],products=[]):
    return pyforms.Form(
        pyforms.Dropdown("node_id", description=u"用户区域", args=nodes, **input_style),
        pyforms.Dropdown("product_id",args=products, description=u"用户资费",  required="required", **input_style),
        pyforms.File("import_file", description=u"用户数据文件",  required="required", **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>立即导入</b>", **button_style),
        title=u"用户数据导入",
        action="/bus/member/import"
)

user_import_vform = dataform.Form(
        dataform.Item("realname", rules.not_null,description=u"用户姓名" ),
        dataform.Item("idcard", rules.len_of(0,32), description=u"证件号码"),
        dataform.Item("mobile", rules.len_of(0,32),description=u"用户手机号码"),
        dataform.Item("address", description=u"用户地址"),
        dataform.Item("account_number",rules.not_null, description=u"用户账号"),
        dataform.Item("password",rules.not_null,description=u"用户密码"),
        dataform.Item("begin_date", rules.is_date,description=u"开通日期"),
        dataform.Item("expire_date", rules.is_date,description=u"过期日期"),
        dataform.Item("balance",rules.is_rmb,description=u"用户余额"),
        dataform.Item("time_length",description=u"用户时长"),
        dataform.Item("flow_length",description=u"用户流量"),
        title="import"
)

account_next_form = pyforms.Form(
        pyforms.Hidden("product_id", description=u"资费"),
        pyforms.Hidden("old_expire", description=u""),
        pyforms.Hidden("account_number", description=u"用户账号"),
        pyforms.Textbox("months",rules.is_number, description=u"月数(包月有效)",value=0, required="required", **input_style),
        pyforms.Textbox("fee_value",rules.is_rmb, description=u"缴费金额",value=0,  required="required", **input_style),
        pyforms.Textbox("expire_date", rules.is_date,description=u"过期日期",  required="required", **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户续费",
        action="/bus/account/next"
)

account_charge_form = pyforms.Form(
        pyforms.Hidden("account_number", description=u"用户账号",  required="required", **input_style),
        pyforms.Textbox("fee_value",rules.is_rmb, description=u"缴费金额", value=0, required="required", **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户充值",
        action="/bus/account/charge"
)

account_cancel_form = pyforms.Form(
        pyforms.Hidden("account_number", description=u"用户账号",  required="required", **input_style),
        pyforms.Textbox("fee_value",rules.is_rmb, description=u"退费金额",  required="required", **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户销户",
        action="/bus/account/cancel"
)


def account_change_form(products=[]):
    return pyforms.Form(
        pyforms.Hidden("account_number", description=u"用户账号",  required="required", **input_style),
        pyforms.Dropdown("product_id",args=products, description=u"资费",  required="required", **input_style),
        pyforms.Textbox("add_value",rules.is_rmb, description=u"缴费金额",required="required",value="0", **input_style),
        pyforms.Textbox("back_value",rules.is_rmb, description=u"退费金额",required="required",value="0", **input_style),
        pyforms.Textbox("expire_date", rules.is_date,description=u"过期日期",value="0000-00-00",**input_style),
        pyforms.Textbox("balance",rules.is_rmb,description=u"用户变更后余额",value="0.00",**input_style),
        pyforms.Textbox("time_length",description=u"用户时长(小时)",value="0.00",**input_style),
        pyforms.Textbox("flow_length",description=u"用户流量(MB)",value="0",**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户资费变更",
        action="/bus/account/change"
)

def member_update_form(nodes=[]):
    return pyforms.Form(
        pyforms.Hidden("account_number", description=u"用户账号"),
        pyforms.Hidden("member_id", description=u"mid",  required="required", **input_style),
        pyforms.Textbox("realname", rules.len_of(2,32), description=u"用户姓名", required="required",**input_style),
        pyforms.Textbox("member_name", description=u"自助服务用户名", readonly="readonly",**input_style),
        pyforms.Password("new_password", rules.len_of(0,128),value="", description=u"自助服务密码(留空不修改)", **input_style),
        pyforms.Textbox("email", rules.len_of(0,128), description=u"电子邮箱", **input_style),
        pyforms.Textbox("idcard", rules.len_of(0,32), description=u"证件号码", **input_style),
        pyforms.Textbox("mobile", rules.len_of(0,32),description=u"用户手机号码", **input_style),
        pyforms.Textbox("address", description=u"用户地址",hr=True, **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户基本信息修改",
        action="/bus/member/update"
    )


def account_update_form():
    return pyforms.Form(
        pyforms.Textbox("account_number", description=u"用户账号",  readonly="readonly", **input_style),
        pyforms.Textbox("ip_address", description=u"用户IP地址",**input_style),
        pyforms.Hidden("install_address", description=u"用户安装地址",**input_style),
        pyforms.Textbox("new_password", description=u"认证密码(留空不修改)", **input_style),
        pyforms.Textbox("user_concur_number",rules.is_number, description=u"用户并发数",  required="required", **input_style),
        pyforms.Dropdown("bind_mac",  args=boolean.items(), description=u"是否绑定MAC",**input_style),
        pyforms.Dropdown("bind_vlan",  args=boolean.items(),description=u"是否绑定VLAN",**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户变更资料",
        action="/bus/account/update"
    )

card_types = {0:u'资费卡',1:u'余额卡'}
card_states = {0:u'未激活',1:u'已激活',2:u"已使用",3:u"已回收"}

def recharge_card_form(products=[]):
    return pyforms.Form(
        pyforms.Dropdown("card_type",  args=card_types.items(), description=u"充值卡类型",**input_style),
        pyforms.Textbox("batch_no", rules.is_number,maxlength=8,description=u"批次号(年+月+2位序号，如：20150201)",required="required", **input_style),
        pyforms.Dropdown("product_id",args=products, description=u"资费",**input_style),
        pyforms.Textbox("start_no",rules.is_number,maxlength=5, description=u"开始卡号(最大5位)",required="required",**input_style),
        pyforms.Textbox("stop_no", rules.is_number,maxlength=5,description=u"结束卡号(最大5位)",required="required",**input_style),
        pyforms.Textbox("pwd_len", rules.is_number,description=u"密码长度(最大为16)",value=8,**input_style),
        pyforms.Textbox("fee_value",rules.is_rmb, description=u"面值/销售价(元)",value=0,**input_style),
        pyforms.Textbox("months", rules.is_number,description=u"授权时间(月)",readonly="readonly",value=0,**input_style),
        pyforms.Textbox("times",description=u"总时长(小时)",readonly="readonly",value=0,**input_style),
        pyforms.Textbox("flows",description=u"总流量(MB)",readonly="readonly",value=0,**input_style),
        pyforms.Textbox("expire_date",rules.is_date, description=u"过期时间",required="required",**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"充值卡生成",
        action="/card/create"
    )
    







