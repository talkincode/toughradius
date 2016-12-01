#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughradius.common.btforms import dataform
from toughradius.common.btforms import rules
from toughradius.common import utils, apiutils, dispatch
from toughradius.common.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius import models

""" 删除区域节点
"""

@permit.route(r"/api/v1/node/delete")
class NodeDeleteHandler(ApiHandler):
    """ @param: 
        node_id: str
    """

    def get(self):
        self.post()

    def post(self):
        try:
            request = self.parse_form_request()
        except apiutils.SignError, err:
            return self.render_sign_err(err)
        except Exception as err:
            return self.render_parse_err(err)

        try:
            node_id = request.get('node_id')

            if not node_id:
                return self.render_verify_err(u'node_id can not be NULL')

            node = self.db.query(models.TrNode).get(node_id)
            if not node:
                return self.render_verify_err(u'node is not exist')

            if self.db.query(models.TrCustomer).filter_by(node_id=node_id).count() > 0:
                return self.render_verify_err(u'node is already bind by customer,can not be delete')

            self.db.query(models.TrNode).filter_by(id=node_id).delete()
            self.add_oplog(u'API删除区域成功:%s' % utils.safeunicode(node_id))
            self.db.commit()
            return self.render_success(msg=u'API删除区域成功:%s' % node_id)
        except Exception as err:
            return self.render_unknow(err)















