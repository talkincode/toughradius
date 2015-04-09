__author__ = '00755'
# demo.py
#
from i18n.translator import Translator
supported_languages = ['en_EN','zh-CN','th_TH']
# activate thai translations
tr = Translator('', supported_languages,'th_TH')
print tr._('Hello world!')
print tr._('管理首页')

