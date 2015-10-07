#!/usr/bin/env python
# coding=utf-8
from hashlib import md5
import json
import time

from bottle import Bottle

from toughradius.console.websock import websock
from toughradius.console.base import *

__prefix__ = "/api/v1"

app = Bottle()
app.config['__prefix__'] = __prefix__



def mksign(secret, params=[]):
    strs = secret + ''.join(params)
    return md5(strs.encode()).hexdigest()


@app.post('/userUnlock')
def user_unlock(db):
    req_msg = {}
    try:
        req_msg = json.loads(request.body.getvalue())
    except Exception as err:
        log.err('parse params error %s' % str(err))
        return dict(code=1, msg='parse params error')

    account_number = req_msg.get('username')
    ipaddr = req_msg.get('ipaddr')
    macaddr = req_msg.get('macaddr')
    sign = req_msg.get('sign')

    params = [p for p in (account_number, ipaddr, macaddr) if p]
    _sign = mksign(app.config['DEFAULT.secret'], params)
    print _sign
    if sign not in _sign:
        return dict(code=1, msg='sign error')

    if not any([account_number, ipaddr, macaddr]):
        return dict(code=1, msg=u'params [username,ipaddr,macaddr] must have at least one')

    query = db.query(models.SlcRadOnline)
    if account_number:
        query = query.filter(models.SlcRadOnline.account_number == account_number)

    if ipaddr:
        query = query.filter(models.SlcRadOnline.framed_ipaddr == ipaddr)

    if macaddr:
        query = query.filter(models.SlcRadOnline.mac_addr == macaddr)

    online = query.first()
    if not online:
        return dict(code=0, msg='online user not find')

    def disconn():
        websock.invoke_admin("coa_request",
                             nas_addr=online.nas_addr,
                             acct_session_id=online.acct_session_id,
                             message_type='disconnect'
                             )

    def is_ok():
        return db.query(models.SlcRadOnline).filter_by(acct_session_id=online.acct_session_id).count() == 0

    disconn()
    count = 0
    while count < 3:
        if is_ok():
            return dict(code=0, msg='success')
        else:
            disconn()
            time.sleep(0.1)
            count += 1

    if not is_ok():
        return dict(code=1, msg='unlock user fail,retry time %s' % count)

    return dict(
        code=0,
        msg='success'
    )