#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from tablib import Dataset
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.admin import card_forms
from toughradius.console.base import *
import bottle
import datetime
from sqlalchemy import func

__prefix__ = "/card"

app = Bottle()
app.config['__prefix__'] = __prefix__


@app.get('/calc',apply=auth_opr)
def card_calc(db, render):
    product_id = request.params.get('product_id')
    product = db.query(models.SlcRadProduct).get(product_id)
    #预付费包月
    if product and product.product_policy == 0:
        return dict(code=0,data={
            'fee_value' : utils.fen2yuan(product.fee_price),
            'months' : 1
        })
    # 买断包月 买断时长 买断流量
    elif product and product.product_policy in (2,3,5):
        return dict(code=0,data={
            'fee_value' : utils.fen2yuan(product.fee_price),
            'months' : product.fee_months,
            'times' : utils.sec2hour(product.fee_times),
            'flows' : utils.kb2mb(product.fee_flows)
        })
    else:
        return dict(code=1,data=u"不支持的资费")
    
@app.route('/list',apply=auth_opr,method=['GET','POST'])
@app.post('/export',apply=auth_opr)
def card_list(db, render):
    product_id = request.params.get('product_id')
    card_type = request.params.get('card_type') 
    card_status = request.params.get('card_status')
    batch_no = request.params.get('batch_no')
    query_begin_time = request.params.get('query_begin_time')
    query_end_time = request.params.get('query_end_time')
    
    _query = db.query(models.SlcRechargerCard)
    if product_id and card_type == '0':
        _query = _query.filter(models.SlcRechargerCard.product_id==product_id)
    if card_type:
        _query = _query.filter(models.SlcRechargerCard.card_type==card_type)
    if batch_no:
        _query = _query.filter(models.SlcRechargerCard.batch_no==batch_no)
    if card_status:
        _query = _query.filter(models.SlcRechargerCard.card_status==card_status)
    if query_begin_time:
        _query = _query.filter(models.SlcRechargerCard.create_time >= query_begin_time+' 00:00:00')
    if query_end_time:
        _query = _query.filter(models.SlcRechargerCard.create_time <= query_end_time+' 23:59:59')
    
    products = db.query(models.SlcRadProduct).filter(
        models.SlcRadProduct.product_status == 0,
        models.SlcRadProduct.product_policy.in_([0,2,3,5])
    )
    if request.path == '/list':
        print "total:",_query.count()
        return render("card_list", 
            page_data = get_page_data(_query),
            card_types = card_forms.card_types,
            card_states = card_forms.card_states,
            products = products,
            colors = {0:'',1:'class="success"',2:'class="warning"',3:'class="danger"'},
            **request.params
        )
    elif request.path == '/export':
        data = Dataset()
        data.append((
            u'批次号',u'充值卡号',u'充值卡密码',u'充值卡类型',u'状态',
            u'资费id', u'面值/售价',u"授权月数",u"授权时长(小时)",u"授权流量(MB)",u"过期时间",u'创建时间'
         ))
        print "total:",_query.count()
        for i in _query:
            data.append((
                i.batch_no, i.card_number, utils.decrypt(i.card_passwd),card_forms.card_types[i.card_type],
                card_forms.card_states[i.card_status],get_product_name(db,i.product_id),utils.fen2yuan(i.fee_value),
                i.months,utils.sec2hour(i.times),utils.kb2mb(i.flows),i.expire_date,i.create_time
            ))
        name = u"RADIUS-CARD-" + datetime.datetime.now().strftime("%Y%m%d-%H%M%S") + ".xls"
        return export_file(name,data)
        
    
permit.add_route("%s/list"%__prefix__,u"充值卡管理",u"系统管理",is_menu=True,order=7)
permit.add_route("%s/export"%__prefix__,u"充值卡导出",u"系统管理",order=7.01)

@app.get('/create',apply=auth_opr)
def card_create(db, render):
    products = [ (n.id,n.product_name) for n in db.query(models.SlcRadProduct).filter(
        models.SlcRadProduct.product_status == 0,
        models.SlcRadProduct.product_policy.in_([0,2,3,5])
    )]
    batch_no = datetime.datetime.now().strftime("%Y%m%d")
    form = card_forms.recharge_card_form(products)
    form.batch_no.set_value(batch_no)
    return render("card_form",form=form)
    
@app.post('/create',apply=auth_opr)
def card_create(db, render):
    def gencardpwd(clen=8):
        r = list('1234567890abcdefghijklmnopqrstuvwxyz')
        rg = utils.random_generator
        return utils.encrypt(''.join([rg.choice(r) for _ in range(clen)]))
        
    products = [ (n.id,n.product_name) for n in db.query(models.SlcRadProduct).filter(
        models.SlcRadProduct.product_status == 0,
        models.SlcRadProduct.product_policy.in_([0,2,3,5])
    )]
    form = card_forms.recharge_card_form(products)
    if not form.validates(source=request.forms):
        return render("card_form",form=form)
    card_type = int(form.d.card_type)
    batch_no = form.d.batch_no
    if len(batch_no) != 8:
        return render("card_form",form=form,msg=u"批次号必须是8位数字")
    
    pwd_len = int(form.d.pwd_len)
    if pwd_len > 16:
        pwd_len = 16

    start_card = int(form.d.start_no)
    stop_card = int(form.d.stop_no)
    if start_card > stop_card:
        return render("card_form",form=form,msg=u"开始卡号不能大于结束卡号")
    
    if form.d.expire_date < utils.get_currdate():
        return render("card_form",form=form,msg=u"过期时间不能小于今天")
    
    fee_value = utils.yuan2fen(form.d.fee_value)
    if fee_value == 0 and card_type == 1:
        return render("card_form",form=form,msg=u"不能发行余额为0的余额卡")
    
    for _card in range(start_card,stop_card+1):
        card_number = "%s%s"%(batch_no,str(_card).zfill(5))
        card_obj = models.SlcRechargerCard()
        card_obj.batch_no = batch_no
        card_obj.card_number = card_number
        card_obj.card_passwd = gencardpwd(pwd_len)
        card_obj.card_type = card_type
        card_obj.card_status = 0
        card_obj.product_id = card_type==0 and form.d.product_id or -1
        card_obj.fee_value = fee_value
        card_obj.months = card_type==0 and int(form.d.months) or 0
        card_obj.times = card_type==0 and utils.hour2sec(form.d.times) or 0
        card_obj.flows = card_type==0 and utils.mb2kb(form.d.flows) or 0
        card_obj.expire_date = form.d.expire_date
        card_obj.create_time = utils.get_currtime()
        db.add(card_obj)
    
    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)生成批次[%s]的[%s]'%(get_cookie("username"),batch_no,card_forms.card_types[card_type])
    db.add(ops_log)
    db.commit()
    path = "%s/list?card_type=%s&query_begin_time=%s"%(__prefix__,card_type,utils.get_currdate())
    if form.d.product_id:
        path = "%s&product_id=%s"%(path,form.d.product_id)
    redirect(path)

permit.add_route("%s/create"%__prefix__,u"充值卡生成",u"系统管理",order=7.02)

@app.get('/active',apply=auth_opr)
def card_active(db, render):
    card_id = request.params.get("card_id")
    if not card_id:
        return dict(code=0,msg=u"非法的访问")
    card = db.query(models.SlcRechargerCard).get(card_id)
    if not card:
        return dict(code=0,msg=u"充值卡不存在")
    card.card_status = 1
    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)激活充值卡[%s]'%(get_cookie("username"),card.card_number)
    db.add(ops_log)
    db.commit()
    return dict(code=0,msg=u"激活成功，充值卡已可使用")
    

permit.add_route("%s/active"%__prefix__,u"充值卡激活",u"系统管理",order=7.03)


@app.get('/recycle',apply=auth_opr)
def card_recycle(db, render):
    card_id = request.params.get("card_id")
    if not card_id:
        return dict(code=0,msg=u"非法的访问")
    card = db.query(models.SlcRechargerCard).get(card_id)
    if not card:
        return dict(code=0,msg=u"充值卡不存在")
    card.card_status = 3
    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)回收充值卡[%s]'%(get_cookie("username"),card.card_number)
    db.add(ops_log)
    db.commit()
    return dict(code=0,msg=u"回收成功，充值卡已不可使用")
    

permit.add_route("%s/recycle"%__prefix__,u"充值卡回收",u"系统管理",order=7.04)