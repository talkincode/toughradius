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
from toughradius.common.permit import permit
from toughradius.common import utils
from collections import deque
from toughradius import models
from toughradius import settings 
from toughradius.common import tools
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


@permit.route(r"/admin/cache/clean")
class CacheClearHandler(BaseHandler):

    def get(self):
        self.cache.clean()
        self.render_json(msg=u"刷新缓存完成")

@permit.route(r"/admin/trace/clean")
class TraceClearHandler(BaseHandler):

    def get(self):
        self.logtrace.clean()
        self.render_json(msg=u"刷新系统消息缓存完成")


@permit.route(r"/admin/dashboard", u"控制面板", settings.MenuSys, order=1.0000, is_menu=True, is_open=False)
class DashboardHandler(BaseHandler):

    def cache_rate(self):
        rate = decimal.Decimal(self.cache.hit_total * 100.0 /(self.cache.get_total+0.0001))
        return str(rate.quantize(decimal.Decimal('1.00')))

    def get_disk_use(self):
        def bb2gb(ik):
            _kb = decimal.Decimal(ik or 0)
            _mb = _kb / decimal.Decimal(1000*1000*1000)
            return str(_mb.quantize(decimal.Decimal('1.00')))
        disks = [ (p,psutil.disk_usage(p)) for p in [a.mountpoint for a in psutil.disk_partitions()] ]
        dstrs = [ "%s %sG/%sG, used %s%%"%(p,bb2gb(d.used),bb2gb(d.total),d.percent) for p,d in disks]
        return '； '.join(dstrs)

    @cyclone.web.authenticated
    def get(self):
        sys_uuid = tools.get_sys_uuid()
        cpuuse = psutil.cpu_percent(interval=None, percpu=True)
        memuse = psutil.virtual_memory()
        diskuse = self.get_disk_use()
        online_count = self.db.query(models.TrOnline.id).count()
        user_total = self.db.query(models.TrAccount.account_number).filter_by(status=1).count()
        self.render("index.html",config=self.settings.config,
            cpuuse=cpuuse,memuse=memuse,diskuse=diskuse,
            online_count=online_count,
            user_total=user_total,
            sys_uuid=sys_uuid,
            cache_rate=self.cache_rate)


class ComplexEncoder(json.JSONEncoder):
    def default(self, obj):
        if type(obj) == deque:
            return [i for i in obj]
        return json.JSONEncoder.default(self, obj)

@permit.route(r"/admin/dashboard/msgstat", u"消息统计", settings.MenuSys, order=1.0001, is_menu=False)
class MsgStatHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        resp = json.dumps(self.cache.get(RADIUS_STATCACHE_KEY), cls=ComplexEncoder,ensure_ascii=False)
        self.write(resp)


@permit.route(r"/admin/dashboard/restart", u"重启服务", settings.MenuSys, order=1.0004, is_menu=False)
class RestartHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        return self.render_json(**execute("supervisorctl restart all && supervisorctl status all"))


@permit.route(r"/admin/dashboard/update", u"更新系统状态", settings.MenuSys, order=1.0002, is_menu=False)
class UpdateHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        return self.render_json(**execute("supervisorctl status all"))


@permit.route(r"/admin/dashboard/upgrade", u"升级系统版本", settings.MenuSys, order=1.0003, is_menu=False)
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


@permit.route(r"/admin/dashboard/onlinestat", u"在线用户统计", settings.MenuSys, order=1.0004, is_menu=False)
class OnlineStatHandler(BaseHandler):

    @cyclone.web.authenticated
    def get(self):
        olstat = self.cache.get(ONLINE_STATCACHE_KEY) or []
        self.render_json(code=0, data=[{'name':u"所有区域",'data': olstat}])


@permit.route(r"/admin/dashboard/flowstat", u"在线用户统计", settings.MenuSys, order=1.0005, is_menu=False)
class FlowStatHandler(BaseHandler):


    def sizedesc(self,ik):
        _kb = decimal.Decimal(ik or 0)
        _mb = _kb / decimal.Decimal(1024)
        return str(_mb.quantize(decimal.Decimal('1.000')))

    @cyclone.web.authenticated
    def get(self):
        flow_stat = self.cache.get(FLOW_STATCACHE_KEY) or {}
        _idata = [(_time,float(self.sizedesc(bb))) for _time,bb in flow_stat.get('input_stat',[]) if bb > 0][-512:]
        _odata = [(_time,float(self.sizedesc(bb))) for _time,bb in flow_stat.get('output_stat',[]) if bb > 0][-512:]
        in_data = {"name": u"上行流量", "data": _idata}
        out_data = {"name": u"下行流量", "data": _odata}

        self.render_json(code=0, data=[in_data, out_data])

