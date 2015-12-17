#!/usr/bin/env python
#coding=utf-8

from toughradius.console.libs import pyforms

not_null = pyforms.notnull
is_not_empty = pyforms.regexp('.+', u"不允许为空")
is_date = pyforms.regexp('(\d{4})-(\d{2}-(\d\d))', u"日期格式:yyyy-MM-dd")
is_email = pyforms.regexp('[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$', u"email格式,比如name@domain.com")
is_chars = pyforms.regexp("^[A-Za-z]+$", u"必须是英文字符串")
is_alphanum = lambda x: pyforms.regexp("^[A-Za-z0-9]{%s}$" % x, u"必须是长度为%s的数字字母组合" % x)
is_alphanum2 = lambda x, y: pyforms.regexp("^[A-Za-z0-9]{%s,%s}$" % (x, y), u"必须是长度为%s到%s的数字字母组合" % (x, y))
is_number = pyforms.regexp("^[0-9]*$", u"必须是数字")
is_number2 = pyforms.regexp("^[1-9]\d*$",u'必须是大于0的正整数')
is_number3 = pyforms.regexp('^(([1-9]\d*)|0)(\.\d{1,3})?$', u"支持包含(最大3位)小数点 xx.xxxxx")
is_numberOboveZore = pyforms.regexp("^\\d+$",u"必须为大于等于0的整数")
is_cn = pyforms.regexp("^[\u4e00-\u9fa5],{0,}$", u"必须是汉字")
is_url = pyforms.regexp('[a-zA-z]+://[^\s]*', u"url格式 xxxx://xxx")
is_phone = pyforms.regexp('^(\(\d{3,4}\)|\d{3,4}-)?\d{7,8}$', u"固定电话号码格式：0000-00000000")
is_idcard = pyforms.regexp('^\d{15}$|^\d{18}$|^\d{17}[Xx]$', u"身份证号码格式")
is_ip = pyforms.regexp("(^$)|\d+\.\d+\.\d+\.\d+", u"ip格式：xxx.xxx.xxx.xxx")
is_rmb = pyforms.regexp('^(([1-9]\d*)|0)(\.\d{1,2})?$', u"人民币金额 xx.xx")
len_of = lambda x, y: pyforms.regexp("[\s\S]{%s,%s}$" % (x, y), u"长度必须为%s到%s" % (x, y))
is_alphanum3 = lambda x, y: pyforms.regexp("^[A-Za-z0-9\_\-]{%s,%s}$" % (x, y), u"必须是长度为%s到%s的数字字母与下划线组合" % (x, y))
is_period = pyforms.regexp("(^$)|^([01][0-9]|2[0-3]):[0-5][0-9]-([01][0-9]|2[0-3]):[0-5][0-9]$",u"时间段，hh:mm-hh:mm,支持跨天，如 19:00-09:20")
is_telephone = pyforms.regexp("^1[0-9]{10}$", u"必须是手机号码")
is_time = pyforms.regexp('(\d{4})-(\d{2}-(\d\d))\s([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]', u"时间格式:yyyy-MM-dd hh:mm:ss")
is_time_hm = pyforms.regexp('^([01][0-9]|2[0-3]):[0-5][0-9]$', u"时间格式: hh:mm")
input_style = {"class": "form-control"}
button_style = {"class": "btn btn-primary"}
button_style_block = {"class": "btn btn-block"}





if __name__ == "__main__":
    print is_period.valid("") == True