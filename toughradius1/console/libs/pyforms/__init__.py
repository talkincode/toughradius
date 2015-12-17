#coding:utf-8
import sys,os
import copy
import re
import itertools
import net

"""basic from web.py: makes web apps (http://webpy.org)"""

__version__ = "0.37"
__author__ = [
    "Aaron Swartz <me@aaronsw.com>",
    "Anand Chitipothu <anandology@gmail.com>"
]
__license__ = "public domain"
__contributors__ = "see https://github.com/webpy/webpy"

class Storage(dict):
    def __getattr__(self, key): 
        try:
            return self[key]
        except KeyError, k:
            raise AttributeError, k
    
    def __setattr__(self, key, value): 
        self[key] = value
    
    def __delattr__(self, key):
        try:
            del self[key]
        except KeyError, k:
            raise AttributeError, k
    
    def __repr__(self):     
        return '<Storage ' + dict.__repr__(self) + '>'

storage = Storage

def attrget(obj, attr, value=None):
    try:
        if hasattr(obj, 'has_key') and obj.has_key(attr): 
            return obj[attr]
    except TypeError:
        # Handle the case where has_key takes different number of arguments.
        # This is the case with Model objects on appengine. See #134
        pass
    if hasattr(obj, attr):
        return getattr(obj, attr)
    return value

def safeunicode(obj, encoding='utf-8'):

    t = type(obj)
    if t is unicode:
        return obj
    elif t is str:
        return obj.decode(encoding)
    elif t in [int, float, bool]:
        return unicode(obj)
    elif hasattr(obj, '__unicode__') or isinstance(obj, unicode):
        return unicode(obj)
    else:
        return str(obj).decode(encoding)    

def safestr(obj, encoding='utf-8'):

    if isinstance(obj, unicode):
        return obj.encode(encoding)
    elif isinstance(obj, str):
        return obj
    elif hasattr(obj, 'next'): # iterator
        return itertools.imap(safestr, obj)
    else:
        return str(obj)

def autoassign(self, locals):

    for (key, value) in locals.iteritems():
        if key == 'self': 
            continue
        setattr(self, key, value)        


class Form(object):

    def __init__(self, *inputs, **kw):
        self.inputs = inputs
        self.valid = True
        self.note = None
        self.validators = kw.pop('validators', [])
        self.action = kw.pop("action","")
        self.title = kw.pop("title","")
        self.method = kw.pop("method","post")
        self.onsubmit = kw.pop("onsubmit","")
        self.error = None

    def __call__(self, x=None):
        o = copy.deepcopy(self)
        if x: o.validates(x)
        return o
    
    def render(self):
        out = ''
        out += self.rendernote(self.note)
        out += '<table class="formtab table table-bordered">\n'
        out += '<thead ><tr class=active><th>%s</th><th class=rtd><a class="btn"\
               href="javascript:history.go(-1);">%s</a></th></tr></thead>\n'%(self.title, net.websafe("返回"))
        
        for i in self.inputs:
            html = safeunicode(i.pre) + i.render() + self.rendernote(i.note) + safeunicode(i.post)
            if i.is_hidden():
                out += '    <tr style="display: none;"><td></td><td>%s</td></tr>\n' % (html)
            else:
                out += '    <tr><td>%s</td><td>%s</td></tr>\n' % ( net.websafe(i.description),html)
        if self.error:
             out += '    <tr><td colspan=2>%s</td></tr>\n' % ( self.rendernote(self.error))
        out += "</table>"
        return out
        
    def render_css(self): 
        out = [] 
        out.append(self.rendernote(self.note)) 
        for i in self.inputs:
            out.append('    <div class="form-group">\n')
            if not i.is_hidden():
                out.append('        <label class="col-sm-4 control-label" id="lab_%s" for="%s">%s</label>\n' % (i.id,i.id, net.websafe(i.description)))
            out.append(i.pre)
            out.append('        <div class="col-sm-6">\n')
            out.append("        %s\n"%i.render())
            out.append('        </div>\n')
            if i.help:
                out.append('    <a id="%s_help" href="javascript:void(0);" data-container="body" data-toggle="popover" data-trigger="focus" data-placement="top" data-content="%s">\n'%(i.id,i.help))
                out.append('    <span class="input-help glyphicon glyphicon-question-sign"></span></a>\n')
            out.append(self.rendernote(i.note))
            out.append(i.post) 
            out.append('    </div>\n')
            if i.hr:
                out.append("<hr/>\n")
        return ''.join(out) 
        
    def rendernote(self, note):
        if note: return '<span class="wrong">%s</span>' % net.websafe(note)
        else: return ""
    
    def validates(self, source=None, _validate=True, **kw):
        source = source or kw
        out = True
        for i in self.inputs:
            v = attrget(source, i.name)
            if _validate:
                out = i.validate(v) and out
            else:
                i.set_value(v)
        if _validate:
            out = out and self._validate(source)
            self.valid = out
        return out

    def _validate(self, value):
        self.value = value
        for v in self.validators:
            if not v.valid(value):
                self.note = v.msg
                return False
        return True

    def fill(self, source=None, **kw):
        return self.validates(source, _validate=False, **kw)
    
    def __getitem__(self, i):
        for x in self.inputs:
            if x.name == i: return x
        raise KeyError, i

    def __getattr__(self, name):
        # don't interfere with deepcopy
        inputs = self.__dict__.get('inputs') or []
        for x in inputs:
            if x.name == name: return x
        raise AttributeError, name
    
    def get(self, i, default=None):
        try:
            return self[i]
        except KeyError:
            return default
            
    def _get_d(self): #@@ should really be form.attr, no?
        return storage([(i.name, i.get_value()) for i in self.inputs])
    d = property(_get_d)

class Input(object):
    def __init__(self, name, *validators, **attrs):
        self.name = name
        self.validators = validators
        self.attrs = attrs = AttributeList(attrs)
        
        self.description = attrs.pop('description', name)
        self.help = attrs.pop("help","")
        self.value = attrs.pop('value', None)
        self.pre = attrs.pop('pre', "")
        self.post = attrs.pop('post', "")
        self.hr = attrs.pop('hr',False)
        self.note = None
        
        self.id = attrs.setdefault('id', self.get_default_id())
        
        if 'class_' in attrs:
            attrs['class'] = attrs['class_']
            del attrs['class_']

        attrs['placeholder'] = self.description
        for vd in self.validators:
            attrs['placeholder'] += ", "+vd.msg



        
    def is_hidden(self):
        return False
        
    def get_type(self):
        raise NotImplementedError
        
    def get_default_id(self):
        return self.name

    def validate(self, value):
        self.set_value(value)

        for v in self.validators:
            if not v.valid(value):
                self.note = v.msg
                return False
        return True

    def set_value(self, value):
        self.value = value

    def get_value(self):
        return self.value

    def render(self):
        attrs = self.attrs.copy()
        attrs['type'] = self.get_type()
        if self.value is not None:
            attrs['value'] = self.value
        attrs['name'] = self.name
        return '<input %s/>' % attrs

    def rendernote(self, note):
        if note: return '<strong class="wrong">%s</strong>' % net.websafe(note)
        else: return ""
        
    def addatts(self):
        # add leading space for backward-compatibility
        return " " + str(self.attrs)

class AttributeList(dict):

    def copy(self):
        return AttributeList(self)
        
    def __str__(self):
        return " ".join(['%s="%s"' % (k, net.websafe(v)) for k, v in self.items()])
        
    def __repr__(self):
        return '<attrs: %s>' % repr(str(self))

class Textbox(Input):
     
    def get_type(self):
        return 'text'

class Password(Input):

    def get_type(self):
        return 'password'

class Textarea(Input):

    def render(self):
        attrs = self.attrs.copy()
        attrs['name'] = self.name
        value = net.websafe(self.value or '')
        return '<textarea %s>%s</textarea>' % (attrs, value)

class Dropdown(Input):

    def __init__(self, name, args, *validators, **attrs):
        self.args = args
        super(Dropdown, self).__init__(name, *validators, **attrs)

    def render(self):
        attrs = self.attrs.copy()
        attrs['name'] = self.name
        
        x = '<select %s>\n' % attrs
        
        for arg in self.args:
            x += self._render_option(arg)

        x += '        </select>\n'
        return x

    def _render_option(self, arg, indent='  '):
        if isinstance(arg, (tuple, list)):
            value, desc= arg
        else:
            value, desc = arg, arg 

        if self.value == value or (isinstance(self.value, list) and value in self.value):
            select_p = ' selected="selected"'
        else:
            select_p = ''
        return indent + '       <option%s value="%s">%s</option>\n' % (select_p, net.websafe(value), net.websafe(desc))
        

class GroupedDropdown(Dropdown):

    def __init__(self, name, args, *validators, **attrs):
        self.args = args
        super(Dropdown, self).__init__(name, *validators, **attrs)

    def render(self):
        attrs = self.attrs.copy()
        attrs['name'] = self.name
        
        x = '   <select %s>\n' % attrs
        
        for label, options in self.args:
            x += '  <optgroup label="%s">\n' % net.websafe(label)
            for arg in options:
                x += self._render_option(arg, indent = '    ')
            x +=  ' </optgroup>\n'
            
        x += '  </select>\n'
        return x

class Radio(Input):
    def __init__(self, name, args, *validators, **attrs):
        self.args = args
        super(Radio, self).__init__(name, *validators, **attrs)

    def render(self):
        x = '<span>'
        for arg in self.args:
            if isinstance(arg, (tuple, list)):
                value, desc= arg
            else:
                value, desc = arg, arg 
            attrs = self.attrs.copy()
            attrs['name'] = self.name
            attrs['type'] = 'radio'
            attrs['value'] = value
            if self.value == value:
                attrs['checked'] = 'checked'
            x += '<input %s/> %s' % (attrs, net.websafe(desc))
        x += '</span>'
        return x

class Checkbox(Input):

    def __init__(self, name, *validators, **attrs):
        self.checked = attrs.pop('checked', False)
        Input.__init__(self, name, *validators, **attrs)
        
    def get_default_id(self):
        value = safestr(self.value or "")
        return self.name + '_' + value.replace(' ', '_')

    def render(self):
        attrs = self.attrs.copy()
        attrs['type'] = 'checkbox'
        attrs['name'] = self.name
        attrs['value'] = self.value

        if self.checked:
            attrs['checked'] = 'checked'            
        return '<input %s/>' % attrs

    def set_value(self, value):
        self.checked = bool(value)

    def get_value(self):
        return self.checked

class Button(Input):

    def __init__(self, name, *validators, **attrs):
        super(Button, self).__init__(name, *validators, **attrs)
        self.description = ""

    def render(self):
        attrs = self.attrs.copy()
        attrs['name'] = self.name
        if self.value is not None:
            attrs['value'] = self.value
        html = attrs.pop('html', None) or net.websafe(self.name)
        return '<button %s>%s</button>' % (attrs, html)

class Hidden(Input):

    def is_hidden(self):
        return True
        
    def get_type(self):
        return 'hidden'

class File(Input):

    def get_type(self):
        return 'file'
    
class Validator:
    def __deepcopy__(self, memo): return copy.copy(self)
    def __init__(self, msg, test, jstest=None): autoassign(self, locals())
    def valid(self, value): 
        try: return self.test(value)
        except: return False

notnull = Validator("Required", bool)

class regexp(Validator):
    def __init__(self, rexp, msg):
        self.rexp = re.compile(rexp)
        self.msg = msg
    
    def valid(self, value):
        return bool(self.rexp.match(value))

if __name__ == "__main__":
    import doctest
    doctest.testmod()
