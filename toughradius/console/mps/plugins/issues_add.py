#coding:utf-8

from cyclone.util import ObjectDict
from toughradius.console import models

__name__ = 'issues_query'

def test(data, msg=None, db=None,**kwargs):
    return data.strip() == '5' or data.strip()  in (u"工单申请")

def respond(data, msg=None,db=None,config=None,mpsapi=None,**kwargs):
    member = db.query(models.SlcMember).filter(
        models.SlcMember.weixin_id==msg.fromuser).first()

    if not member:
        return u"您当前还未绑定账号"


    articles = []
    article=ObjectDict()
    article.title= u"工单申请"
    article.description = ''
    article.url = "%s/issues/add?openid=%s&member_id=%s" % (
        config.get('mps','server_base'),
        msg.fromuser,
        member.member_id
    )
    article.picurl = '%s/static/img/mps/oprgds.jpg' % config.get('mps','server_base')
    articles.append(article)
    return articles