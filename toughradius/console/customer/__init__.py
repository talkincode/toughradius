from toughradius.console.libs import i18n
from io import open
#
# # use the Translator class directly:
# #tr = i18n.Translator('../toughradius/console/lang.yml', language='th', fallback='en')
#tr = i18n.Translator('../toughradius/console/lang.yml')
# # or use the load_translator() function:
tr = i18n.load_translator('../toughradius/console/customer/lang.yml')
# #tr.language = 'th'
# # tr.fallback = ''
#_ = tr.t