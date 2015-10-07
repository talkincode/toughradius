#!/usr/bin/env python
# coding=utf-8
from bottle import Bottle
import bottle

from toughradius.console.websock import websock
from toughradius.console.base import *
from toughradius.console.admin import param_forms

__prefix__ = "/param"

app = Bottle()
app.config['__prefix__'] = __prefix__


###############################################################################
# param config
###############################################################################

@app.get('/', apply=auth_opr)
def param(db, render):
    active = request.params.get("active","syscfg")
    sys_form = param_forms.sys_form()
    serv_form = param_forms.serv_form()
    notify_form = param_forms.notify_form()
    mail_form = param_forms.mail_form()
    rad_form = param_forms.rad_form()
    fparam = {}
    for p in db.query(models.SlcParam):
        fparam[p.param_name] = p.param_value

    for form in (sys_form, serv_form, notify_form, mail_form, rad_form):
        form.fill(fparam)

    return render("sys_param",
                  active=active,
                  sys_form=sys_form,
                  serv_form=serv_form,
                  notify_form=notify_form,
                  mail_form=mail_form,
                  rad_form=rad_form)


@app.post('/update', apply=auth_opr)
def param_update(db, render):
    params = db.query(models.SlcParam)
    active = request.params.get("active", "syscfg")
    for param_name in request.forms:
        if param_name in ("active", "submit"):
            continue

        param = db.query(models.SlcParam).filter_by(param_name=param_name).first()
        if not param:
            param = models.SlcParam()
            param.param_name = param_name
            param.param_value = request.forms.get(param_name)
            db.add(param)
        else:
            param.param_value = request.forms.get(param_name)

    ops_log = models.SlcRadOperateLog()
    ops_log.operator_name = get_cookie("username")
    ops_log.operate_ip = get_cookie("login_ip")
    ops_log.operate_time = utils.get_currtime()
    ops_log.operate_desc = u'操作员(%s)修改参数' % (get_cookie("username"))
    db.add(ops_log)
    db.commit()

    if "radiusd_address" in request.forms:
        websock.reconnect(
            request.forms.get('radiusd_address'),
            request.forms.get('radiusd_admin_port'),
        )

    if "is_debug" in request.forms:
        is_debug = request.forms.get('is_debug')
        bottle.debug(is_debug == '1')
        websock.update_cache("is_debug", is_debug=is_debug)
    if "reject_delay" in request.forms:
        websock.update_cache("reject_delay", reject_delay=request.forms.get('reject_delay'))

    websock.update_cache("param")
    redirect("/param?active=%s"% active)


permit.add_route("/param", u"系统参数管理", MenuSys, is_menu=True, order=0.0001)
permit.add_route("/param/update", u"系统参数修改", MenuSys, is_menu=False, order=0.0002, is_open=False)
