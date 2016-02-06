#!/usr/bin/env python
# coding:utf-8
import os
import json
import subprocess
import datetime
import time
import os.path
import cyclone.auth
import cyclone.escape
import cyclone.web
from toughradius.manage.base import BaseHandler
from txradius import statistics
from toughlib.permit import permit
from toughlib import utils
from collections import deque
from toughradius.manage import models
from toughradius.manage.settings import * 
import psutil

##############################################################################
# basic
##############################################################################

class ToughError(Exception):
    def __init__(self, message):
        self.message = message


def run_command(command, raise_error_on_fail=False, shell=True, env=None):
    _result = dict(code=0)
    run_env = os.environ.copy()
    if env: run_env.update(env)
    proc = subprocess.Popen(command, shell=shell,
                            stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                            env=run_env)
    stdout, stderr = proc.communicate('through stdin to stdout')
    result = proc.returncode, stdout, stderr
    if proc.returncode > 0 and raise_error_on_fail:
        error_string = "* Could not run command (return code= %s)\n" % proc.returncode
        error_string += "* Error was:\n%s\n" % (stderr.strip())
        error_string += "* Command was:\n%s\n" % command
        error_string += "* Output was:\n%s\n" % (stdout.strip())
        if proc.returncode == 127:  # File not found, lets print path
            path = os.getenv("PATH")
            error_string += "Check if y/our path is correct: %s" % path
        raise ToughError(error_string)
    else:
        return result


def warp_html(code, value):
    _value = value.replace("\n", "<br>")
    _value = _value.replace("RUNNING", "<strong><font color=green>RUNNING</font></strong>")
    _value = _value.replace("STARTING", "<strong><font color='#CC9900'>STARTING</font></strong>")
    _value = _value.replace("FATAL", "<strong><font color=red>FATAL</font></strong>")
    if code > 0:
        _value = '<font color="#CC0000">%s</font>' % _value
    return _value


def execute(cmd):
    try:
        rcode, stdout, stderr = run_command(cmd, True)
        return dict(value=warp_html(rcode, (stdout or stderr)))
    except ToughError, err:
        import traceback
        traceback.print_exc()
        return dict(value=warp_html(1, err.message))


##############################################################################
# web handler
##############################################################################

@permit.route(r"/admin/dashboard", u"控制面板", MenuSys, order=1.0000, is_menu=True, is_open=False)
class DashboardHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        cpuuse = psutil.cpu_percent(interval=None, percpu=True)
        memuse = psutil.virtual_memory()
        online_count = self.db.query(models.TrOnline.id).count()
        user_total = self.db.query(models.TrAccount.account_number).filter_by(status=1).count()
        self.render("index.html",config=self.settings.config,
            cpuuse=cpuuse,memuse=memuse,online_count=online_count,user_total=user_total)


class ComplexEncoder(json.JSONEncoder):
    def default(self, obj):
        if type(obj) == deque:
            return [i for i in obj]
        return json.JSONEncoder.default(self, obj)

@permit.route(r"/admin/dashboard/msgstat", u"消息统计", MenuSys, order=1.0001, is_menu=False)
class MsgStatHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        resp = json.dumps(self.cache.get(radius_statcache_key), cls=ComplexEncoder,ensure_ascii=False)
        self.write(resp)

@permit.route(r"/admin/dashboard/msgstat/reset", u"消息统计重置", MenuSys, order=1.0001, is_menu=False)
class MsgStatResetHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        self.cache.delete(radius_statcache_key)
        resp = json.dumps(statistics.MessageStat(), cls=ComplexEncoder,ensure_ascii=False)
        self.write(resp)

@permit.route(r"/admin/dashboard/restart", u"重启服务", MenuSys, order=1.0004, is_menu=False)
class RestartHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        return self.render_json(**execute("supervisorctl restart all && supervisorctl status all"))


@permit.route(r"/admin/dashboard/update", u"更新系统状态", MenuSys, order=1.0002, is_menu=False)
class UpdateHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        return self.render_json(**execute("supervisorctl status all"))


@permit.route(r"/admin/dashboard/upgrade", u"升级系统版本", MenuSys, order=1.0003, is_menu=False)
class UpgradeHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        release = self.get_argument("release")
        cmd1 = "cd /opt/toughradius"
        cmd2 = "git fetch origin %s && git checkout %s && git submodule update --recursive" % (release, release)
        cmd3 = "supervisorctl restart all"
        return self.render_json(**execute("%s && %s && %s" % (cmd1, cmd2, cmd3)))

def default_start_end():
    day_code = datetime.datetime.now().strftime("%Y-%m-%d")
    begin = datetime.datetime.strptime("%s 00:00:00" % day_code, "%Y-%m-%d %H:%M:%S")
    end = datetime.datetime.strptime("%s 23:59:59" % day_code, "%Y-%m-%d %H:%M:%S")
    return time.mktime(begin.timetuple()), time.mktime(end.timetuple())

@permit.route(r"/admin/dashboard/onlinestat", u"在线用户统计", MenuSys, order=1.0004, is_menu=False)
class OnlineStatHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        node_id = self.get_argument('node_id',None)
        day_code = self.get_argument('day_code',None)
        opr_nodes = self.get_opr_nodes()
        if not day_code:
            day_code = utils.get_currdate()
        begin = datetime.datetime.strptime("%s 00:00:00" % day_code, "%Y-%m-%d %H:%M:%S")
        end = datetime.datetime.strptime("%s 23:59:59" % day_code, "%Y-%m-%d %H:%M:%S")
        begin_time, end_time = time.mktime(begin.timetuple()), time.mktime(end.timetuple())
        _query = self.db.query(models.TrOnlineStat)

        if node_id:
            _query = _query.filter(models.TrOnlineStat.node_id == node_id)
        else:
            _query = _query.filter(models.TrOnlineStat.node_id.in_([i.id for i in opr_nodes]))

        _query = _query.filter(
            models.TrOnlineStat.stat_time >= begin_time,
            models.TrOnlineStat.stat_time <= end_time,
        )
        _data = [(q.stat_time * 1000, q.total) for q in _query]
        self.render_json(code=0, data=[{'data': _data}])


@permit.route(r"/admin/dashboard/flowstat", u"在线用户统计", MenuSys, order=1.0005, is_menu=False)
class FlowStatHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        node_id = self.get_argument('node_id',None)
        day_code = self.get_argument('day_code',None)
        opr_nodes = self.get_opr_nodes()
        if not day_code:
            day_code = utils.get_currdate()
        begin = datetime.datetime.strptime("%s 00:00:00" % day_code, "%Y-%m-%d %H:%M:%S")
        end = datetime.datetime.strptime("%s 23:59:59" % day_code, "%Y-%m-%d %H:%M:%S")
        begin_time, end_time = time.mktime(begin.timetuple()), time.mktime(end.timetuple())
        _query = self.db.query(models.TrFlowStat)

        if node_id:
            _query = _query.filter(models.TrFlowStat.node_id == node_id)
        else:
            _query = _query.filter(models.TrFlowStat.node_id.in_([i.id for i in opr_nodes]))

        _query = _query.filter(
            models.TrFlowStat.stat_time >= begin_time,
            models.TrFlowStat.stat_time <= end_time,
        )

        in_data = {"name": u"上行流量", "data": []}
        out_data = {"name": u"下行流量", "data": []}

        for q in _query:
            _stat_time = q.stat_time * 1000
            in_data['data'].append([_stat_time, float(utils.kb2mb(q.input_total))])
            out_data['data'].append([_stat_time, float(utils.kb2mb(q.output_total))])

        self.render_json(code=0, data=[in_data, out_data])



