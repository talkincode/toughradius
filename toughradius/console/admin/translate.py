__author__ = '00755'
#!/usr/bin/env python
#coding:utf-8
import sys

import py

from toughradius.tools.i18n.translator import Translator


# set the root of the project to the directory containing this file
ROOT = py.path.local(__file__).dirpath()
LANGUAGES = ['TH' , 'EN']

tr = Translator(ROOT, LANGUAGES, 'zh_Hans_CN')
_ = tr._
ngettext = tr.ngettext

if __name__ == '__main__':
    tr.cmdline(sys.argv)
