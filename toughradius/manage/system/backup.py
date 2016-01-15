#!/usr/bin/env python
# coding:utf-8
import os
import os.path
import cyclone.auth
import cyclone.escape
import cyclone.web
from toughlib import utils,dispatch,logger
from toughradius.manage.base import BaseHandler
from toughlib.permit import permit
from toughradius.manage.settings import * 

@permit.route(r"/admin/backup", u"数据备份管理", MenuSys, order=5.0001, is_menu=True)
class BackupHandler(BaseHandler):
    @cyclone.web.authenticated
    def get(self):
        backup_path = self.settings.config.database.backup_path
        try:
            if not os.path.exists(backup_path):
                os.makedirs(backup_path)
        except:
            pass
        flist = os.listdir(backup_path)
        flist.sort(reverse=True)
        return self.render("backup_db.html", backups=flist[:30], backup_path=backup_path)

@permit.route(r"/admin/backup/dump", u"备份数据", MenuSys, order=5.0002)
class DumpHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        backup_path = self.settings.config.database.backup_path
        backup_file = "toughradius_db_%s.json.gz" % utils.gen_backep_id()
        try:
            self.db_backup.dumpdb(os.path.join(backup_path, backup_file))
            return self.render_json(code=0, msg="backup done!")
        except Exception as err:
            dispatch.pub(logger.EVENT_EXCEPTION,err)
            return self.render_json(code=1, msg="backup fail! %s" % (err))

@permit.route(r"/admin/backup/restore", u"恢复数据", MenuSys, order=5.0003)
class RestoreHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        backup_path = self.settings.config.database.backup_path
        backup_file = "toughradius_db_%s.before_restore.json.gz" % utils.gen_backep_id()
        rebakfs = self.get_argument("bakfs")
        try:
            self.db_backup.dumpdb(os.path.join(backup_path, backup_file))
            self.db_backup.restoredb(os.path.join(backup_path, rebakfs))
            return self.render_json(code=0, msg="restore done!")
        except Exception as err:
            dispatch.pub(logger.EVENT_EXCEPTION,err)
            return self.render_json(code=1, msg="restore fail! %s" % (err))


@permit.route(r"/admin/backup/delete", u"删除数据", MenuSys, order=5.0004)
class DeleteHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        backup_path = self.settings.config.database.backup_path
        bakfs = self.get_argument("bakfs")
        try:
            os.remove(os.path.join(backup_path, bakfs))
            return self.render_json(code=0, msg="delete done!")
        except Exception as err:
            dispatch.pub(logger.EVENT_EXCEPTION,err)
            return self.render_json(code=1, msg="delete fail! %s" % (err))


@permit.route(r"/admin/backup/upload", u"上传数据", MenuSys, order=5.0004)
class UploadHandler(BaseHandler):
    @cyclone.web.authenticated
    def post(self):
        try:
            f = self.request.files['Filedata'][0]
            save_path = os.path.join(self.settings.config.database.backup_path, f['filename'])
            tf = open(save_path, 'wb')
            tf.write(f['body'])
            tf.close()
            self.write("upload success")
        except Exception as err:
            dispatch.pub(logger.EVENT_EXCEPTION,err)
            self.write("upload fail %s" % str(err))










