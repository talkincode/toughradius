#coding=utf-8
import re
from cyclone.util import ObjectDict

__name__ = 'help'

p = re.compile(r'^h(\s*)$|^help(\s*)$|(.*)menu_help')


def test(data, msg=None, db=None,config=None,**kwargs):
    """
    >>> test('h')
    True
    >>> test('help')
    True
    >>> test('event:CLICK:menu_help')
    True
    >>> test('H')
    True
    >>> test('Help')
    True
    >>> test("He")
    False
    >>> test('h ')
    True
    >>> test('* ')
    True
    >>> test('? ')
    True
    """
    if p.match(data.lower().strip()) or data.lower().strip() in (u'帮助',u'帮助信息'):
        return True
    return False


def respond(data, msg=None,db=None,**kwargs):
    result_str =  u'''发送h可显示帮助提示。" \n
发送以下关键字可以快捷为您服务: \n
1、账号查询 \n
2、账单查询 \n
3、在线订购 \n
4、工单查询 \n
5、工单申请 \n
更多需求可直接发送内容给我们。
'''
    return result_str

if __name__ == 'help':
    import doctest
    doctest.testmod()
