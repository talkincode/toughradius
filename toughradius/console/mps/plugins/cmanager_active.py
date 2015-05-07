#coding=utf-8

from cyclone.util import ObjectDict
from toughradius.console.libs import utils
from toughradius.console import models
from twisted.python import log

__name__ = 'cmanager_active'

def test(data, msg=None,db=None,config=None,mpsapi=None,**kwargs):
    return data.strip().startswith("cmauth")

def respond(data, msg=None,db=None,config=None,mpsapi=None,**kwargs):
    active_code = data.strip()[6:]
    try:
        cmanager = db.query(models.SlcCustomerManager).filter(
            models.SlcCustomerManager.active_code==active_code).first()
        if cmanager:
            if cmanager.active_status == 1:
                return u"客户经理已经激活"
            else:
                cmanager.manager_openid = msg.fromuser
                cmanager.active_status = 1
                cmanager.active_time = utils.get_currtime()
                db.commit()
                return u"客户经理激活成功，当用户扫描你的二维码的时候，你将第一时间收到通知。当你的用户发送消息时，你将第一时间收到。"
        else:
            return u"激活码错误"
    except:
        db.rollback()
        log.err("cmanager_active plugin process error")
        return u"服务器忙，请稍后再试"
    finally:
        db.close()
