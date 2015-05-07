#!/usr/bin/env python
#coding=utf-8
from toughradius.console.mps import base
from cyclone.util import ObjectDict
from twisted.python import log
from toughradius.console.mps.issues_forms import issues_add_form


class AddIssuesHandler(base.BaseHandler):


    def get(self):
        openid=self.get_argument('openid')
        form = issues_add_form()
        form.fill(openid=openid)
        self,render("mps_issues_add.html",form=form)


    def post(self):
        pass
