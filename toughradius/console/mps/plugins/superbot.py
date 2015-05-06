#coding:utf-8

from toughradius.console.mps import mpsmsg
from toughradius.console.models import SlcMember

def respond(data, msg=None, db=None,config=None,**kwargs):
    user = db.query(SlcMember).filter(SlcMember.weixin_id==msg.fromuser).first()
    if user and user.manager_code:
        return dict(
            msg_type=mpsmsg.MSG_TYPE_CUSTOMER,
            kfaccount='kf%s@cdcatv'%user.manager_code
        )
    else:
        return dict(
            msg_type=mpsmsg.MSG_TYPE_CUSTOMER,
            kfaccount='kf10000@cdcatv'
        ) 