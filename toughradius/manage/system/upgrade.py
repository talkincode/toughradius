#!/usr/bin/env python
# coding:utf-8
import os
import cyclone.auth
import cyclone.escape
import cyclone.web
import traceback
import json
from urllib import urlencode
from toughlib import utils,dispatch,logger
from twisted.internet import defer
from toughradius.manage.base import BaseHandler,authenticated
from toughlib.permit import permit
from toughradius import models
from toughradius.manage.settings import * 
from toughradius.common import tools
from cyclone import httpclient
import toughradius
import treq
import zipfile
import shutil

type_descs = {
    'dev' : '<span class="label label-info">开发版</span>',
    'stable' : '<span class="label label-success">稳定版</span>',
    'oem' : '<span class="label label-default">OEM 版</span>', 
    'patch' : '<span class="label label-primary">升级补丁</span>', 
}

@permit.route(r"/admin/upgrade", u"系统升级管理", MenuSys, order=10.0000, is_menu=True)
class UpgradeMetadataHandler(BaseHandler):

    def type_desc(self,typestr):
        return type_descs.get(str(typestr))

    @authenticated
    def get(self):
        try:
            self.render("upgrade.html")
        except Exception as err:
            logger.exception(err)
            self.render_error(msg=repr(err))


@permit.route(r"/admin/upgrade/upload", u"上传版本升级", MenuSys, order=10.0004)
class UpgradeUploadHandler(BaseHandler):
    @authenticated
    def post(self):
        try:
            f = self.request.files['Filedata'][0]
            save_path = "/tmp/{0}".format(f['filename'])
            tf = open(save_path, 'wb')
            tf.write(f['body'])
            tf.close()

            try:
                shutil.rmtree("/tmp/toughradius-upgrades/toughradius")
            except:
                pass

            zipFile = zipfile.ZipFile(save_path)
            zipFile.extractall('/tmp/toughradius-upgrades')
            udst = os.path.dirname(toughradius.__file__)
            tools.copydir("/tmp/toughradius-upgrades/toughradius",udst)

            try:
                ctlfile = "/tmp/toughradius-upgrades/toughradius/radiusctl"
                if os.path.exists("ctlfile"):
                    shutil.copy(ctlfile, "/opt/toughradius/radiusctl")
            except:
                traceback.print_exc()


            self.write(u"升级完成,请重启所有服务")
        except Exception as err:
            logger.error(err)
            self.write(u"升级失败 %s" % utils.safeunicode(err))








