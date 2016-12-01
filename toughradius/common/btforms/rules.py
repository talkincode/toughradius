#!/usr/bin/env python
#coding=utf-8

from toughradius.common import btforms

not_null = btforms.notnull
is_not_empty = btforms.regexp('.+', u"不允许为空")
is_date = btforms.regexp('(\d{4})-(\d{2}-(\d\d))', u"日期格式:yyyy-MM-dd")
is_email = btforms.regexp('[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$', u"email格式,比如name@domain.com")
is_chars = btforms.regexp("^[A-Za-z]+$", u"必须是英文字符串")
is_alphanum = lambda x: btforms.regexp("^[A-Za-z0-9]{%s}$" % x, u"必须是长度为%s的数字字母组合" % x)
is_alphanum2 = lambda x, y: btforms.regexp("^[A-Za-z0-9]{%s,%s}$" % (x, y), u"必须是长度为%s到%s的数字字母组合" % (x, y))
is_number = btforms.regexp("^[0-9]*$", u"必须是数字")
is_number2 = btforms.regexp("^[1-9]\d*$",u'必须是大于0的正整数')
is_number3 = btforms.regexp('^(([1-9]\d*)|0)(\.\d{1,3})?$', u"支持包含(最大3位)小数点 xx.xxxxx")
is_numberOboveZore = btforms.regexp("^\\d+$",u"必须为大于等于0的整数")
is_cn = btforms.regexp("^[\u4e00-\u9fa5],{0,}$", u"必须是汉字")
is_url = btforms.regexp('[a-zA-z]+://[^\s]*', u"url格式 xxxx://xxx")
is_phone = btforms.regexp('^(\(\d{3,4}\)|\d{3,4}-)?\d{7,8}$', u"固定电话号码格式：0000-00000000")
is_idcard = btforms.regexp('^\d{15}$|^\d{18}$|^\d{17}[Xx]$', u"身份证号码格式")
is_ip = btforms.regexp("(^$)|\d+\.\d+\.\d+\.\d+", u"ip格式：xxx.xxx.xxx.xxx")
is_rmb = btforms.regexp('^(([1-9]\d*)|0)(\.\d{1,2})?$', u"人民币金额 xx.xx")
len_of = lambda x, y: btforms.regexp("[\s\S]{%s,%s}$" % (x, y), u"长度必须为%s到%s" % (x, y))
is_alphanum3 = lambda x, y: btforms.regexp("^[A-Za-z0-9\_\-]{%s,%s}$" % (x, y), u"必须是长度为%s到%s的数字字母与下划线组合" % (x, y))
is_period = btforms.regexp("(^$)|^([01][0-9]|2[0-3]):[0-5][0-9]-([01][0-9]|2[0-3]):[0-5][0-9]$",u"时间段，hh:mm-hh:mm,支持跨天，如 19:00-09:20")
is_telephone = btforms.regexp("^1[0-9]{10}$", u"必须是手机号码")
is_time = btforms.regexp('(\d{4})-(\d{2}-(\d\d))\s([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]', u"时间格式:yyyy-MM-dd hh:mm:ss")
is_time_hm = btforms.regexp('^([01][0-9]|2[0-3]):[0-5][0-9]$', u"时间格式: hh:mm")
input_style = {"class": "form-control"}
button_style = {"class": "btn btn-primary"}
button_style_block = {"class": "btn btn-block"}




if __name__ == "__main__":
    print is_number.valid('') 