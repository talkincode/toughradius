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
from toughradius.console.libs.mpsapi import mpsapi
from toughradius.console.base import *
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

permit.add_route("%s/list"%__prefix__,u"用户工单查询",u"营业管理",is_menu=True,order=5.1)

@app.get('/detail',apply=auth_opr)
def issues_detail(db):   
    issues_id = request.params.get("issues_id")
    issues = db.query(models.SlcIssues).get(issues_id)    
    issues_flows = db.query(models.SlcIssues).filter_by(issues_id=issues_id)
    return render("bus_issues_detail",issues=issues,issues_flows=issues_flows)

permit.add_route("%s/detail"%__prefix__,u"用户工单详情",u"营业管理",is_menu=False,order=5.11)


@app.get('/process',apply=auth_opr)
def issues_process(db):   
    pass

@app.get('/process',apply=auth_opr)
def issues_process_post(db):   
    pass









