#coding=utf-8
import requests
from toughradius.console.libs import utils
from toughradius.console import models
from cyclone.util import ObjectDict
from hashlib import md5

__name__ = 'subscribe'

def rand_number(db):
    r = ['0','1','2','3','4','5','6','7','8','9']
    rg = utils.random_generator
    def random_number():
        _num = ''.join([rg.choice(r) for _ in range(7)])
        if db.query(models.SlcMember).filter_by(member_name=_num).count() > 0:
            return random_account()
        else:
            return _num
    return random_number()


def test(data, msg=None, bot=None, db=None,**kwargs):
    return data.strip().startswith('event:subscribe')

def respond(data, msg=None,db=None,config=None,mpsapi=None,**kwargs):

    def get_wlannotify():
        articles =[]
        article1=ObjectDict()
        article1.title=u"无线上网"
        article1.description=u"当您已经连接Wi-Fi信号时，点击即可免费联入网络。"
        article1.url = "%s/mplogin?mp_openid=%s&product_id=%s&node_id=%s" % (
            config.get('mps','server_base'),
            msg.fromuser,
            config.get('mps','wlan_product_id'),
            config.get('mps','wlan_node_id')
        )
        article1.picurl = '%s/static/img/wlan.jpg' % config.get('mps','server_base')
        articles.append(article1)
        return articles

    member = db.query(models.SlcMember).filter(
        models.SlcMember.weixin_id==msg.fromuser).first()

    manager_code = None

    if data.strip().startswith('event:subscribe:qrscene_cmqr_'):
        manager_code = data[31:]

    if data.strip().startswith('event:subscribe:qrscene_wlan_cmqr'):
        manager_code = data[36:]
        return get_wlannotify()

    if data.strip().startswith('event:subscribe:qrscene_portal'):
        return get_wlannotify()

    result_str =  u'''感谢您的关注! 发送h可显示帮助提示。" \n
发送以下关键字可以快捷为您服务: \n
1、账号查询 \n
2、账单查询 \n
3、在线订购 \n
4、工单查询 \n
更多需求可直接发送内容给我们。
'''

    if member:
        if manager_code:
            member.manager_code = manager_code
            db.commit()
            result_str += u'\n工号为%s的客户经理将竭诚为您服务。'%manager_code
    else:
        result_str += u'\n绑定您的的账号，即可在线办理各种业务。'

    return result_str