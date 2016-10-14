#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib import utils, apiutils, dispatch
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models
from toughlib.btforms import dataform
from toughlib.btforms import rules

nas_update_vform = dataform.Form(
    dataform.Item("dns_name", rules.len_of(0, 128), description=u"DNS域名"),
    dataform.Item("bas_name", rules.len_of(0, 64), description=u"接入设备名称", default=u"new bas"),
    dataform.Item("bas_secret", rules.is_alphanum2(4, 32), description=u"共享秘钥"),
    dataform.Item("vendor_id", description=u"接入设备类型", default=0),
    dataform.Item("coa_port", rules.is_number, description=u"授权端口", default=3799),
    title="api nas update"
)


@permit.route(r"/api/v1/nas/update")
class NasUpdateHandler(ApiHandler):

    def get(self):
        self.post()

    def post(self):

        form = nas_update_vform()
        try:
            request = self.parse_form_request()
            if not form.validates(**request):
                return self.render_verify_err(form.errors)
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            if not request.get('ip_addr'):
                return self.render_verify_err(msg=u"nas ip_addr can not be NULL")

            bas = self.db.query(models.TrBas).filter_by(ip_addr=request.get("ip_addr")).first()

            if not bas:
                return self.render_verify_err(msg=u"nas is not exists")

            dns_name = form.d.dns_name
            bas_name = form.d.bas_name
            vendor_id = form.d.vendor_id
            bas_secret = form.d.bas_secret
            coa_port = form.d.coa_port

            if dns_name:
                bas.dns_name = dns_name

            if bas_name:
                bas.bas_name = bas_name

            if vendor_id:
                bas.vendor_id = vendor_id

            if bas_secret:
                bas.bas_secret = bas_secret

            if coa_port:
                bas.coa_port = coa_port
            self.db.commit()
            self.render_success(msg=u'API更新BAS:%s' % request.get('ip_addr'))
        except Exception, err:
            return self.render_unknow(err)



