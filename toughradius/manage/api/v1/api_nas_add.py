#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib import utils, apiutils, dispatch
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib.apiutils import apistatus
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models

nas_add_vform = dataform.Form(
    dataform.Item("ip_addr", rules.is_ip, description=u"接入设备地址"),
    dataform.Item("dns_name", rules.len_of(0, 128), description=u"DNS域名"),
    dataform.Item("bas_name", rules.len_of(0, 64), description=u"接入设备名称",default="new bas"),
    dataform.Item("bas_secret", rules.is_alphanum2(4, 32), description=u"共享秘钥"),
    dataform.Item("vendor_id", description=u"接入设备类型",default=0),
    dataform.Item("coa_port", rules.is_number, description=u"授权端口", default=3799),
    dataform.Item("time_type", description=u"时间类型",default=0),
    title="api nas add"
)

@permit.route(r"/api/v1/nas/add")
class NasAddHandler(ApiHandler):

    def get(self):
        self.post()

    def post(self):
        form = nas_add_vform()
        try:
            request = self.parse_form_request()
            if not form.validates(**request):
                raise Exception(form.errors)
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            if  not any([request.get('ip_addr'),request.get("dns_name")]):
                raise ValueError("ip_addr, dns_name required one")
        except Exception, err:
            return self.render_verify_err(err)

        try:
            if self.db.query(models.TrBas.id).filter_by(ip_addr=request.get("ip_addr")).count() > 0:
                return self.render_verify_err(msg=u"nas already exists")

            bas = models.TrBas()
            bas.ip_addr = request.get("ip_addr")
            bas.dns_name = request.get("dns_name")
            bas.bas_name = request.get("bas_name","new bas")
            bas.time_type = request.get("time_type",0)
            bas.vendor_id = request.get("vendor_id",0)
            bas.bas_secret = request.get("bas_secret")
            bas.coa_port = request.get("coa_port",3799)
            self.db.add(bas)
            self.db.commit()
            self.render_success()

        except Exception, err:
            return self.render_unknow(err)

