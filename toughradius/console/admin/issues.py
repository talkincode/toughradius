#!/usr/bin/env python
#coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from sqlalchemy import func
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.base import *
from toughradius.console.admin import issues_forms
import bottle
import datetime

__prefix__ = "/issues"

app = Bottle()
app.config['__prefix__'] = __prefix__
render = functools.partial(Render.render_app,app)

###############################################################################
# issues manage        
###############################################################################

@app.route('/list',apply=auth_opr,method=['GET','POST'])
def issues_list(db):   
    manager_code = request.params.get('manager_code')
    account_number = request.params.get('account_number')
    issues_type = request.params.get('issues_type')
    status = request.params.get('status')

    _query = db.query(models.SlcIssues)
    if manager_code:
        _query = _query.filter_by(manager_code=manager_code)
    if account_number:
        _query = _query.filter_by(account_number=account_number)
    if issues_type:
        _query = _query.filter_by(issues_type=issues_type)
    if status:
        _query = _query.filter_by(status=status)

    return render("bus_issues_list",page_data = get_page_data(_query),**request.params)



@app.get('/detail',apply=auth_opr)
def issues_detail(db):   
    issues_id = request.params.get("issues_id")
    issues = db.query(models.SlcIssues).get(issues_id)    
    issues_flows = db.query(models.SlcIssuesFlow).filter_by(issues_id=issues_id)

    form = issues_forms.issues_process_form()

    return render("bus_issues_detail",issues=issues,issues_flows=issues_flows, form=form)


@app.get('/add', apply=auth_opr)
def issues_add(db):
    oprs = [(o.operator_name, o.operator_name) for o in db.query(models.SlcOperator)]
    return render("base_form", form=issues_forms.issues_add_form(oprs))


@app.post('/add', apply=auth_opr)
def issues_add_post(db):
    oprs = [(o.operator_name, o.operator_name) for o in db.query(models.SlcOperator)]
    form = issues_forms.issues_add_form(oprs)
    if not form.validates(source=request.forms):
        return render("base_form", form=form)
    if db.query(models.SlcRadAccount).filter_by(account_number=form.d.account_number).count() == 0:
        return render("base_form", form=form,msg=u"用户账号不存在")

    issues = models.SlcIssues()
    issues.account_number = form.d.account_number
    issues.issues_type = form.d.issues_type
    issues.content = form.d.content
    issues.assign_operator = get_cookie("username")
    issues.status = 0
    issues.date_time = utils.get_currtime()

    db.add(issues)
    db.commit()
    redirect("/issues/list")

@app.post('/process',apply=auth_opr)
def issues_process_post(db):   
    pass

@app.get('/delete', apply=auth_opr)
def issues_delete_post(db):
    pass


permit.add_route("%s/list" % __prefix__, u"用户工单查询", u"营业管理", is_menu=True, order=5.1)
permit.add_route("%s/detail" % __prefix__, u"用户工单详情", u"营业管理", is_menu=False, order=5.11)
permit.add_route("%s/add" % __prefix__, u"创建用户工单", u"营业管理", is_menu=False, order=5.12)
permit.add_route("%s/process" % __prefix__, u"处理用户工单", u"营业管理", is_menu=False, order=5.13)
permit.add_route("%s/delete" % __prefix__, u"删除用户工单", u"营业管理", is_menu=False, order=5.14)




