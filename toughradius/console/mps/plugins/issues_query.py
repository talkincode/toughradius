#coding:utf-8

from cyclone.util import ObjectDict
from toughradius.console import models

__name__ = 'issues_query'

def test(data, msg=None, db=None,**kwargs):
    return data.strip() == '4' or data.strip()  in (u"工单查询",u"工单")

def respond(data, msg=None,db=None,config=None,mpsapi=None,**kwargs):
    member = db.query(models.SlcMember).filter(
        models.SlcMember.weixin_id==msg.fromuser).first()

    if not member:
        return u"您当前还未绑定账号"


    articles =[]
    article=ObjectDict()
    article.title= u"我的工单"
    article.description = ''
    article.url = "%s/customer/issues?openid=%s&member_id=%s" % (
        config.get('mps','server_base'),
        msg.fromuser,
        member.member_id
    )
    article.picurl = '%s/static/img/mps/issues_query.jpg' % config.get('mps','server_base')
    articles.append(article)
    return articles