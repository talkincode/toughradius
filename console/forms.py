#coding:utf-8
from libs import pyforms
from libs.pyforms import rules

def param_form(params=[]):
    inputs = []
    for param in params:
        _input = pyforms.Textbox(param.param_name, description=param.param_desc,value=param.param_value, **rules.input_style)
        inputs.append(_input)
    inputs.append(pyforms.Button("submit", type="submit", html=u"<b>提交</b>",**rules.button_style))
    form = pyforms.Form(*inputs,title=u"参数管理",action="/param")
    return form
