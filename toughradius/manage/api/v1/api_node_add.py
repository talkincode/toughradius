#!/usr/bin/env python
#coding=utf-8

from toughradius.common.btforms import dataform
from toughradius.common.btforms import rules
from toughradius.common import utils, apiutils, dispatch
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models

""" 添加区域节点
"""

node_add_vform = dataform.Form(
    dataform.Item("node_name", rules.len_of(2, 32), description=u"区域名称"),
    dataform.Item("node_desc", rules.len_of(2, 128), description=u"区域描述"),
    title=u"node add",
)

@permit.route(r"/api/v1/node/add")
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
            if self.db.query(models.TrNode).filter_by(node_name=form.d.node_name).count() > 0:
                return self.render_verify_err("node name already exists")
        except Exception, err:
            return self.render_verify_err(err)

        try:
            node = models.TrNode()
            node.node_name = form.d.node_name
            node.node_desc = form.d.node_desc or u''
            self.db.add(node)
            self.add_oplog(u'API添加区域成功:%s' % utils.safeunicode(form.d.node_name))
            self.db.commit()
            return self.render_success(msg=u'区域添加成功:%s' % form.d.node_name)
        except Exception as err:
            return self.render_unknow(err)















