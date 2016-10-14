#!/usr/bin/env python
#coding=utf-8

from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib import utils, apiutils, dispatch
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models

""" 修改区域节点
"""

node_add_vform = dataform.Form(
    dataform.Item("node_name", rules.len_of(2, 32), description=u"区域名称"),
    dataform.Item("node_desc", rules.len_of(2, 128), description=u"区域描述"),
    title=u"node update",
)

@permit.route(r"/api/v1/node/update")
class NodeAddHandler(ApiHandler):
    """ @param: 
        form
    """

    def get(self):
        self.post()

    def post(self):
        form = node_add_vform()
        try:
            request = self.parse_form_request()
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)
        try:
            if not form.validates(**request):
                return self.render_verify_err(form.errors)
        except Exception, err:
            return self.render_verify_err(err)

        try:
            node_id = request.get('node_id')
            if not node_id:
                return self.render_verify_err(u'node_id can not be NULL')

            node = self.db.query(models.TrNode).get(node_id)

            if form.d.node_name:
                node.node_name = form.d.node_name

            if form.d.node_desc:
                node.node_desc = form.d.node_desc
            self.add_oplog(u'API修改区域成功:%s' % utils.safeunicode(node_id))
            self.db.commit()
            return self.render_success(msg=u'区域修改成功:%s' % node_id)
        except Exception as err:
            return self.render_unknow(err)















