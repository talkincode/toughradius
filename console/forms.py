#coding:utf-8
from libs import pyforms
from libs.pyforms import rules
from libs.pyforms.rules import button_style,input_style

boolean = {0:u"否", 1:u"是"}

def param_form(params=[]):
    inputs = []
    for param in params:
        _input = pyforms.Textbox(param.param_name, description=param.param_desc,value=param.param_value, **input_style)
        inputs.append(_input)
    inputs.append(pyforms.Button("submit", type="submit", html=u"<b>提交</b>",**button_style))
    return pyforms.Form(*inputs,title=u"参数管理",action="/param")

passwd_update_form = pyforms.Form(
    pyforms.Textbox("operator_name", description=u"管理员名", size=32, readonly="readonly", **input_style),
    pyforms.Password("operator_pass", rules.len_of(6, 32), description=u"管理员新密码", size=32,value="", required="required", **input_style),
    pyforms.Password("operator_pass_chk", rules.len_of(6, 32), description=u"确认管理员新密码", size=32,value="", required="required", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"管理密码更新",
    action="/passwd"
)

node_add_form = pyforms.Form(
    pyforms.Textbox("node_name", rules.len_of(2,32), description=u"节点名称",required="required",**input_style),
    pyforms.Textarea("node_desc", rules.len_of(0, 128), description=u"节点描述",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>",**button_style),
    title=u"增加节点",
    action="/node/add"
)

node_update_form = pyforms.Form(
    pyforms.Hidden("id",  description=u"节点编号"),
    pyforms.Textbox("node_name", rules.len_of(2, 32), description=u"节点名称", required="required",**input_style),
    pyforms.Textarea("node_desc", rules.len_of(0, 128), description=u"节点描述", **input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改节点",
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
            '14988' : u'RouterOS'
        }

bas_add_form = pyforms.Form(
    pyforms.Textbox("ip_addr", rules.is_ip, description=u"BAS地址",required="required",**input_style),
    pyforms.Textbox("bas_name", rules.len_of(2,64), description=u"BAS名称",required="required",**input_style),
    pyforms.Textbox("bas_secret", rules.is_alphanum2(4, 32), description=u"共享秘钥", required="required",**input_style),
    pyforms.Dropdown("vendor_id", description=u"BAS类型", args=bastype.items(), required="required",**input_style),
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
    pyforms.Dropdown("time_type", description=u"时间类型", args=timetype.items(), required="required",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改BAS",
    action="/bas/update"
)

product_policy = {0:u'包月',1:u"小时计费"}
product_status_dict = {0:u'正常',1:u"停用"}

product_add_form = pyforms.Form(
    pyforms.Textbox("product_name", rules.len_of(4, 64), description=u"套餐名称",  required="required",**input_style),
    pyforms.Dropdown("product_policy", args=product_policy.items(), description=u"计费策略", required="required",**input_style),
    pyforms.Dropdown("product_status", args=product_status_dict.items(),description=u"套餐状态", required="required", **input_style),
    pyforms.Textbox("domain_name",rules.len_of(0,16),description=u"套餐域",**input_style),
    pyforms.Textbox("fee_price", rules.is_rmb, description=u"套餐价格(包月价/每小时价)(元)", required="required", **input_style),
    pyforms.Textbox("fee_period",rules.is_period,description=u"计费时段(小时计费有效)",**input_style),
    pyforms.Textbox("concur_number", rules.is_numberOboveZore,description=u"并发数控制(0表示不限制)",value="0", **input_style),
    pyforms.Dropdown("bind_mac",  args=boolean.items(), description=u"是否绑定MAC ",**input_style),
    pyforms.Dropdown("bind_vlan",  args=boolean.items(),description=u"是否绑定VLAN ",**input_style),
    pyforms.Textbox("bandwidth_code",  rules.is_alphanum2(0,16),description=u"套餐限速编码 ",**input_style),
    pyforms.Textbox("input_max_limit",  rules.is_number,description=u"最大上行速率 ", **input_style),
    pyforms.Textbox("output_max_limit",  rules.is_number,description=u"最大下行速率 ",**input_style),
    pyforms.Textbox("input_rate_code",  rules.is_alphanum2(0,16),description=u"上行限速编码 ", **input_style),
    pyforms.Textbox("output_rate_code",  rules.is_alphanum2(0,16),description=u"下行限速编码 ",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加套餐",
    action="/product/add"
)

product_update_form = pyforms.Form(
    pyforms.Hidden("id",  description=u"编号"),
    pyforms.Textbox("product_name", rules.len_of(4, 32), description=u"产品套餐名称", required="required",**input_style),
    pyforms.Textbox("product_policy_name", description=u"套餐策略",readonly="readonly", required="required",**input_style),
    pyforms.Dropdown("product_status",args=product_status_dict.items(), description=u"套餐状态", required="required", **input_style),
    pyforms.Textbox("domain_name",rules.len_of(0,16),description=u"套餐域",**input_style),
    pyforms.Textbox("fee_price", rules.is_rmb,description=u"套餐价格(包月价/每小时价)(元)", required="required", **input_style),
    pyforms.Textbox("fee_period", rules.is_period,description=u"计费时段(小时计费有效)",**input_style),
    pyforms.Textbox("concur_number", rules.is_number,description=u"并发数控制(0表示不限制)", **input_style),
    pyforms.Dropdown("bind_mac",  args=boolean.items(), description=u"是否绑定MAC",**input_style),
    pyforms.Dropdown("bind_vlan",  args=boolean.items(),description=u"是否绑定VLAN",**input_style),
    pyforms.Textbox("bandwidth_code",  rules.is_alphanum2(0,16),description=u"套餐限速编码",**input_style),
    pyforms.Textbox("input_max_limit",  rules.is_number,description=u"最大上行速率", **input_style),
    pyforms.Textbox("output_max_limit",  rules.is_number,description=u"最大下行速率",**input_style),
    pyforms.Textbox("input_rate_code",  rules.is_alphanum2(0,16),description=u"上行限速编码", **input_style),
    pyforms.Textbox("output_rate_code",  rules.is_alphanum2(0,16),description=u"下行限速编码",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改套餐",
    action="/product/update"
)


group_add_form = pyforms.Form(
    pyforms.Textbox("group_name", rules.len_of(2,32), description=u"用户组名",required="required",**input_style),
    pyforms.Textbox("group_desc", rules.len_of(2,64), description=u"用户组描述",required="required",**input_style),
    pyforms.Dropdown("bind_mac",  args=boolean.items(),description=u"绑定MAC", required="required",**input_style),
    pyforms.Dropdown("bind_vlan", args=boolean.items(), description=u"绑定VLAN", required="required",**input_style),
    pyforms.Textbox("concur_number", rules.is_number,description=u"并发数",value=0, required="required",**input_style),
    pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
    title=u"增加用户组",
    action="/group/add"
)

group_update_form = pyforms.Form(
    pyforms.Hidden("id",  description=u"编号"),
    pyforms.Textbox("group_name", rules.len_of(2,32), description=u"用户组名",required="required",**input_style),
    pyforms.Textbox("group_desc", rules.len_of(2,64), description=u"用户组描述",required="required",**input_style),
    pyforms.Dropdown("bind_mac",  args=boolean.items(),description=u"绑定MAC", required="required",**input_style),
    pyforms.Dropdown("bind_vlan", args=boolean.items(), description=u"绑定VLAN", required="required",**input_style),
    pyforms.Textbox("concur_number", rules.is_number,description=u"并发数", required="required",**input_style),
    pyforms.Button("submit",  type="submit", html=u"<b>更新</b>", **button_style),
    title=u"修改用户组",
    action="/group/update"
)














