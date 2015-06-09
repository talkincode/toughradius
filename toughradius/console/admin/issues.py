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

###############################################################################
# issues manage        
###############################################################################

@app.route('/list',apply=auth_opr,method=['GET','POST'])
def issues_list(db, render):
    oprs = [('','')]+[(o.operator_name, o.operator_name) for o in db.query(models.SlcOperator)]
    operator_name = request.params.get('operator_name')
    account_number = request.params.get('account_number')
    issues_type = request.params.get('issues_type')
    status = request.params.get('status')

    _query = db.query(models.SlcIssues)
    if operator_name:
        _query = _query.filter_by(assign_operator=operator_name)
    if account_number:
        _query = _query.filter_by(account_number=account_number)
    if issues_type:
        _query = _query.filter_by(issues_type=issues_type)
    if status:
        _query = _query.filter_by(status=status)


    return render("bus_issues_list",page_data = get_page_data(_query),oprs=oprs,**request.params)



@app.get('/detail',apply=auth_opr)
def issues_detail(db, render):
    issues_id = request.params.get("issues_id")
    issues = db.query(models.SlcIssues).get(issues_id)    
    issues_flows = db.query(models.SlcIssuesFlow).filter_by(issues_id=issues_id)

    form = issues_forms.issues_process_form()
    form.issues_id.set_value(issues_id)
    oprs = [(o.operator_name, o.operator_name) for o in db.query(models.SlcOperator)]
    colors = {0: 'label label-default', 1: 'class="label label-info"', 2: 'class="label label-warning"', 3: 'class="label label-danger"',4:'class="label label-success"'}

    return render("bus_issues_detail",issues=issues,issues_flows=issues_flows, form=form, colors=colors,oprs=oprs)


@app.get('/add', apply=auth_opr)
def issues_add(db, render):
    oprs = [(o.operator_name, o.operator_name) for o in db.query(models.SlcOperator)]
    return render("base_form", form=issues_forms.issues_add_form(oprs))


@app.post('/add', apply=auth_opr)
def issues_add_post(db, render):
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
    issues.assign_operator = form.d.assign_operator
    issues.status = 0
    issues.date_time = utils.get_currtime()

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)创建新工单' % (get_cookie("username") )
    db.add(ops_log)

    db.add(issues)
    db.commit()
    redirect("/issues/list")

@app.post('/process',apply=auth_opr)
def issues_process_post(db, render):
    form = issues_forms.issues_process_form()
    if not form.validates(source=request.forms):
        return render("base_form", form=form)

    iflow = models.SlcIssuesFlow()
    iflow.issues_id = form.d.issues_id
    iflow.accept_time = utils.get_currtime()
    iflow.accept_status = form.d.accept_status
    iflow.accept_result = form.d.accept_result
    iflow.operator_name = get_cookie("username")
    db.add(iflow)

    issues = db.query(models.SlcIssues).get(iflow.issues_id)
    issues.status = iflow.accept_status

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)处理工单%s' % (get_cookie("username"),iflow.issues_id)
    db.add(ops_log)

    db.commit()

    redirect("/issues/detail?issues_id=%s"%iflow.issues_id)


@app.post('/assign', apply=auth_opr)
def issues_assign_post(db, render):
    issues_id = request.params.get("issues_id")
    assign_operator = request.params.get("assign_operator")

    if assign_operator == get_cookie("username"):
        redirect("/issues/detail?issues_id=%s" % issues_id)
        return

    issues = db.query(models.SlcIssues).get(issues_id)
    issues.assign_operator = assign_operator

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)转派工单%s给%s' % (get_cookie("username"), issues_id, assign_operator)
    db.add(ops_log)

    db.commit()
    redirect("/issues/detail?issues_id=%s" % issues_id)


@app.get('/delete', apply=auth_opr)
def issues_delete_post(db, render):
    issues_id = request.params.get("issues_id")
    db.query(models.SlcIssues).filter_by(id=issues_id).delete()
    db.query(models.SlcIssuesFlow).filter_by(issues_id=issues_id)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)删除工单%s' % (get_cookie("username"), issues_id)
    db.add(ops_log)

    db.commit()

    redirect("/issues/list")

permit.add_route("%s/list" % __prefix__, u"用户工单管理", u"营业管理", is_menu=True, order=5.1)
permit.add_route("%s/detail" % __prefix__, u"用户工单详情", u"营业管理", is_menu=False, order=5.11)
permit.add_route("%s/add" % __prefix__, u"创建用户工单", u"营业管理", is_menu=False, order=5.12)
permit.add_route("%s/process" % __prefix__, u"处理用户工单", u"营业管理", is_menu=False, order=5.13)
permit.add_route("%s/assign" % __prefix__, u"转派用户工单", u"营业管理", is_menu=False, order=5.14)
permit.add_route("%s/delete" % __prefix__, u"删除用户工单", u"营业管理", is_menu=False, order=5.15)



