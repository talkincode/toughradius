#!/usr/bin/env python
#coding=utf-8
import time

from toughradius.common import utils
from toughradius.common.permit import permit
from toughradius.console.admin.api import api_base
from toughradius.console import models


@permit.route(r"/api/admin")
class AdminhHandler(api_base.ApiHandler):

    def get(self):
        self.post()

    def post(self):
        self.render_json(msg="ok")


