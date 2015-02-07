#coding:utf-8
from libs import pyforms
from libs.pyforms import dataform
from libs.pyforms import rules
from libs.pyforms.rules import button_style,input_style

boolean = {0:u"否", 1:u"是"}

sexopt = {1:u"男",0:u"女"}

member_login_form = pyforms.Form(
    pyforms.Textbox("username", rules.len_of(1, 32), description=u"用户名", size=32,required="required",**input_style),
    pyforms.Password("password", rules.len_of(1,32), description=u"登录密码", size=32, required="required",**input_style),
    pyforms.Button("submit", type="submit", html=u"<b>登陆</b>", **button_style),
    pyforms.Hidden("next",value="/"),
    action="/login",
    title=u"用户登陆"
)

def member_join_form(nodes=[]): 
    return pyforms.Form(
        pyforms.Dropdown("node_id", description=u"区域", args=nodes,required="required", **input_style),
        pyforms.Textbox("realname", rules.len_of(2,32), description=u"用户姓名(必填)", required="required",**input_style),
        pyforms.Dropdown("sex", description=u"性别", args=sexopt.items(),required="required", **input_style),
        pyforms.Textbox("age", rules.is_number, description=u"年龄(必填)", size=3,required="required",**input_style),
        pyforms.Textbox("username", rules.is_alphanum3(6, 32), description=u"用户名(必填)", size=32,required="required",**input_style),
        pyforms.Password("password", rules.len_of(6,32), description=u"登录密码(必填)", size=32, required="required",**input_style),
        pyforms.Textbox("email", rules.is_email, description=u"电子邮箱(必填)", size=64,required="required",**input_style),
        pyforms.Textbox("idcard", rules.len_of(0,32), description=u"证件号码", **input_style),
        pyforms.Textbox("mobile", rules.len_of(0,32),description=u"用户手机号码", **input_style),
        pyforms.Textbox("address", description=u"用户地址",hr=True, **input_style),
        pyforms.Button("submit", type="submit", html=u"<b>注册</b>", **button_style),
        action="/join",
        title=u"用户注册"
    )
    
password_update_form =  pyforms.Form(
        pyforms.Textbox("account_number", description=u"上网账号",  readonly="readonly", **input_style),
        pyforms.Password("old_password",description=u"旧密码(必填)", required="required",**input_style),
        pyforms.Password("new_password", rules.is_alphanum3(6, 32),description=u"新密码(必填)", required="required",**input_style),
        pyforms.Password("new_password2",rules.is_alphanum3(6, 32), description=u"确认新密码(必填)", required="required",**input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"修改密码",
        action="/password/update"
    )


def account_open_form(products=[]):
    return pyforms.Form(
        pyforms.Textbox("recharge_card", description=u"充值卡号", required="required", **input_style),
        pyforms.Password("recharge_pwd", description=u"充值卡密码", required="required", **input_style),
        pyforms.Textbox("account_number", description=u"上网账号",  required="required", **input_style),
        pyforms.Textbox("password", description=u"上网密码", required="required", **input_style),
        pyforms.Dropdown("product_id",args=products, description=u"上网资费",  required="required", **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户自助开户",
        action="/open"
    )

recharge_form =  pyforms.Form(
        pyforms.Textbox("account_number",description=u"上网账号",readonly="readonly", **input_style),
        pyforms.Textbox("recharge_card", description=u"充值卡号", required="required", **input_style),
        pyforms.Password("recharge_pwd", description=u"充值卡密码", required="required", **input_style),
        pyforms.Button("submit",  type="submit", html=u"<b>提交</b>", **button_style),
        title=u"用户自助充值",
        action="/recharge"
    )




