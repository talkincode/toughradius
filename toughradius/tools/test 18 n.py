from toughradius.console.libs import i18n

# use the Translator class directly:
tr = i18n.Translator('foo.yml', language='th', fallback='en')

# or use the load_translator() function:
tr = i18n.load_translator('foo.yml')
tr.language = 'jj'
tr.fallback = 'de'
_ = tr.t

print(_(u'greet'))
print(_(u'mynameis').format(name='Siegfried'))
