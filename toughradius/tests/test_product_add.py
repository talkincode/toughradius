#!/usr/bin/env python
#coding:utf-8
from toughradius.tests.test_base import TestMixin
from twisted.trial import unittest
from twisted.internet import reactor
import sys
import os

class InitdbTestCase(unittest.TestCase,TestMixin):

    def setUp(self):
        self.init_rundir()
        self.init_config()

    def test_add_ppmonth(self):
        pass

    def test_add_pptimes(self):
        pass

    def test_add_bomonth(self):
        pass

    def test_add_botimes(self):
        pass

    def test_add_ppflows(self):
        pass

    def test_add_boflows(self):
        pass

