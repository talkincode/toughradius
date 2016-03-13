#!/usr/bin/env python
#coding:utf-8
from toughradius.tests.test_base import TestMixin
from twisted.trial import unittest
import sys
import os

class NodeAddTestCase(unittest.TestCase,TestMixin):

    def setUp(self):
        self.init_rundir()
        self.init_config()

    def test_add_bas(self):
        req = self.admin_login()
        r = req.post(self.sub_path("/admin/node/add"),dict(node_name="testnode",node_desc=u"测试区域2"))
        assert r.status_code == 200