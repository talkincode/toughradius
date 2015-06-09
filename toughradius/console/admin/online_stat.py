#!/usr/bin/env python
# coding:utf-8

from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import static_file
from bottle import mako_template as render
from tablib import Dataset
from toughradius.console.websock import websock
from toughradius.console import models
from toughradius.console.libs import utils
from toughradius.console.base import *
from toughradius.console.admin import forms
import bottle
import datetime
from sqlalchemy import func

__prefix__ = "/online_stat"

app = Bottle()
app.config['__prefix__'] = __prefix__

###############################################################################
# ops log manage
###############################################################################

def default_start_end():
    day_code = datetime.datetime.now().strftime("%Y-%m-%d")
    begin = datetime.datetime.strptime("%s 00:00:00" % day_code, "%Y-%m-%d %H:%M:%S")
    end = datetime.datetime.strptime("%s 23:59:59" % day_code, "%Y-%m-%d %H:%M:%S")
    return time.mktime(begin.timetuple()), time.mktime(end.timetuple())


@app.get('/', apply=auth_opr)
def online_stat_query(db, render):
    return render(
        "stat_online",
        node_list=get_opr_nodes(db),
        node_id=None,
        day_code=utils.get_currdate()
    )


@app.route('/data', apply=auth_opr, method=['GET', 'POST'])
def online_stat_data(db, render):
    node_id = request.params.get('node_id')
    day_code = request.params.get('day_code')
    opr_nodes = get_opr_nodes(db)
    if not day_code:
        day_code = utils.get_currdate()
    begin = datetime.datetime.strptime("%s 00:00:00" % day_code, "%Y-%m-%d %H:%M:%S")
    end = datetime.datetime.strptime("%s 23:59:59" % day_code, "%Y-%m-%d %H:%M:%S")
    begin_time, end_time = time.mktime(begin.timetuple()), time.mktime(end.timetuple())
    _query = db.query(models.SlcRadOnlineStat)

    if node_id:
        _query = _query.filter(models.SlcRadOnlineStat.node_id == node_id)
    else:
        _query = _query.filter(models.SlcRadOnlineStat.node_id.in_([i.id for i in opr_nodes]))

    _query = _query.filter(
        models.SlcRadOnlineStat.stat_time >= begin_time,
        models.SlcRadOnlineStat.stat_time <= end_time,
    )
    _data = [(q.stat_time * 1000, q.total) for q in _query]
    return dict(code=0, data=[{'data': _data}])


permit.add_route("/online_stat", u"在线用户统计", MenuStat, is_menu=True, order=1)