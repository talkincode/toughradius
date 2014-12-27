#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from bottle import mako_template as render
from hashlib import md5
from tablib import Dataset
import bottle
import models
import forms
import datetime
from base import *

app = Bottle()

@app.route('/member',apply=auth_opr,method=['GET','POST'])
@app.get('/member/export',apply=auth_opr)
def member_query(db):   
    node_id = request.params.get('node_id')
    realname = request.params.get('realname')
    _query = db.query(
            models.SlcMember.realname,
            models.SlcMember.member_id,
            models.SlcMember.mobile,
            models.SlcMember.address,
            models.SlcMember.create_time,
            models.SlcNode.node_name
        ).filter(
            models.SlcNode.id == models.SlcMember.node_id
        )
    if node_id:
        _query = _query.filter(models.SlcMember.node_id == node_id)
    if realname:
        _query = _query.filter(models.SlcMember.realname.like('%'+realname+'%'))


    if request.path == '/member':
        return render("bus_member_list", page_data = get_page_data(_query),
                       node_list=db.query(models.SlcNode),**request.params)
    elif request.path == "/member/export":
        result = _query.all()
        data = Dataset()
        data.append((u'区域',u'姓名', u'联系电话', u'地址', u'创建时间'))
        for i in result:
            data.append((i.node_name, i.realname, i.mobile, i.address,i.create_time))
        name = u"RADIUS-MEMBER-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        with open(u'./static/xls/%s' % name, 'wb') as f:
            f.write(data.xls)
        return static_file(name, root='./static/xls',download=True)    