import sys

import py

from toughradius.tools.i18n.translator import Translator


# set the root of the project to the directory containing this file
ROOT = py.path.local(__file__).dirpath()
LANGUAGES = ['EN', 'TH']

tr = Translator(ROOT, LANGUAGES, 'TH')
_ = tr._
ngettext = tr.ngettext

if __name__ == '__main__':
    tr.cmdline(sys.argv)