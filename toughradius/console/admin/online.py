#!/usr/bin/env python
# coding:utf-8

from bottle import Bottle
from toughradius.console.base import *

__prefix__ = "/online"

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# online manage
###############################################################################

@app.route('/query', apply=auth_opr, method=['GET', 'POST'])
def online_query(db, render):
    node_id = request.params.get('node_id')
    account_number = request.params.get('account_number')
    framed_ipaddr = request.params.get('framed_ipaddr')
    mac_addr = request.params.get('mac_addr')
    nas_addr = request.params.get('nas_addr')
    opr_nodes = get_opr_nodes(db)
    _query = db.query(
        models.SlcRadOnline.id,
        models.SlcRadOnline.account_number,
        models.SlcRadOnline.nas_addr,
        models.SlcRadOnline.acct_session_id,
        models.SlcRadOnline.acct_start_time,
        models.SlcRadOnline.framed_ipaddr,
        models.SlcRadOnline.mac_addr,
        models.SlcRadOnline.nas_port_id,
        models.SlcRadOnline.start_source,
        models.SlcRadOnline.billing_times,
        models.SlcRadOnline.input_total,
        models.SlcRadOnline.output_total,
        models.SlcMember.node_id,
        models.SlcMember.realname
    ).filter(
        models.SlcRadOnline.account_number == models.SlcRadAccount.account_number,
        models.SlcMember.member_id == models.SlcRadAccount.member_id
    )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    else:
        _query = _query.filter(models.SlcMember.node_id.in_([i.id for i in opr_nodes]))

    if account_number:
        _query = _query.filter(models.SlcRadOnline.account_number.like('%' + account_number + '%'))
    if framed_ipaddr:
        _query = _query.filter(models.SlcRadOnline.framed_ipaddr == framed_ipaddr)
    if mac_addr:
        _query = _query.filter(models.SlcRadOnline.mac_addr == mac_addr)
    if nas_addr:
        _query = _query.filter(models.SlcRadOnline.nas_addr == nas_addr)

    _query = _query.order_by(models.SlcRadOnline.acct_start_time.desc())
    return render("ops_online_list", page_data=get_page_data(_query),
                  node_list=opr_nodes,
                  bas_list=db.query(models.SlcRadBas), **request.params)


permit.add_route("/online/query" , u"在线用户查询", u"维护管理", is_menu=True, order=2)