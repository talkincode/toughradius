#!/usr/bin/env python
#coding=UTF8
from toughradius.console.i18n.translator import Translator
supported_languages = ['TH','EN']
# activate italian translations
tr = Translator('./', supported_languages, 'TH')
print tr._('Hello world!')

