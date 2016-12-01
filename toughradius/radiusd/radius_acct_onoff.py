#!/usr/bin/env python
# coding=utf-8

from toughradius.radiusd.radius_basic import  RadiusBasic
from toughradius.common.storage import Storage
from toughradius import models
from toughradius.common import  utils, dispatch, logger
from toughradius.manage.settings import *

class RadiusAcctOnoff(RadiusBasic):

    def __init__(self, dbengine=None,cache=None,aes=None,request=None):
        RadiusBasic.__init__(self, dbengine,cache,aes, request)

    def acctounting(self):
        logger.info('bas accounting onoff success')
