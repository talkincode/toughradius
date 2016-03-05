#!/usr/bin/env python
#coding=utf-8
import time
import traceback
from toughlib.btforms import dataform
from toughlib.btforms import rules
from toughlib import utils, apiutils, dispatch
from toughlib.permit import permit
from toughradius.manage.api.apibase import ApiHandler
from toughradius.manage import models

""" 区域节点管理
"""

@permit.route(r"/api/v1/node/query")
class NodeQueryHandler(ApiHandler):
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
            nodes = self.db.query(models.TrNode)
            if node_id:
                nodes = nodes.filter_by(id=node_id)

            node_datas = []
            for node in nodes:
                node_data = { c.name : getattr(node, c.name) for c in node.__table__.columns}
                node_datas.append(node_data)

            self.render_success(nodes=node_datas)

        except Exception as err:
            return self.render_unknow(err)















