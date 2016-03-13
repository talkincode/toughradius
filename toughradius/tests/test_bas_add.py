#!/usr/bin/env python
#coding:utf-8
from toughradius.tests.test_base import TestMixin
from twisted.trial import unittest
from toughradius.manage.resource import bas_forms
import sys
import os

class BasAddTestCase(unittest.TestCase,TestMixin):

    def setUp(self):
        self.init_rundir()
        self.init_config()

    def test_add_bas(self):
        form = bas_forms.bas_add_form()
        assert form.validates(source=dict(
            ip_addr="192.168.31.153",
            bas_name="stdbas",
            bas_secret="secret",
            vendor_id='0',
            coa_port='3799',
            time_type='0'
        ))
        req = self.admin_login()
        r = req.post(self.sub_path("/admin/bas/add"),form.d)
        assert r.status_code == 200