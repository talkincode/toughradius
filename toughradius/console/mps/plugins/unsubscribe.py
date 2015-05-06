#coding=utf-8
from toughradius.console import models
__name__ = 'unsubscribe'

def test(data, msg=None,db=None,**kwargs):
    return data.strip().startswith('event:unsubscribe')

def respond(data, msg=None,db=None,**kwargs):
    user = db.query(models.SlcMember).filter(
        models.SlcMember.weixin_id==msg.fromuser).first()
    if user:
        user.weixin_id = ''
        db.commit()
    return "bye"