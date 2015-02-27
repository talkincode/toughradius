#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
import argparse,ConfigParser
from datetime import datetime
from ftplib import FTP

def backup(config):
    import sh
    dbname=config.get('database','db')
    ftphost=config.get('database','ftphost')
    ftpport=config.get('database','ftpport')
    ftpuser=config.get('database','ftpuser')
    ftppwd=config.get('database','ftppwd')
    
    bakdir = "/var/toughradius/databak"
    if not os.path.exists(bakdir):
        os.mkdir(bakdir)
        
    now = datetime.now()
    backfile = '%s/%s-backup-%s.gz'%(bakdir,dbname,now.strftime( "%Y%m%d"))
    sh.gzip(sh.mysqldump(u='root',B=dbname),'-cf',_out=backfile)

    if ftphost and '127.0.0.1' not in ftphost:
        ftp=FTP() 
        ftp.set_debuglevel(2)
        ftp.connect(ftphost,ftpport)
        ftp.login(ftpuser,ftppwd)
        ftp.cwd('/')
        bufsize = 1024
        file_handler = open(backfile,'rb')
        ftp.storbinary('STOR %s' % os.path.basename(backfile),file_handler,bufsize)
        ftp.set_debuglevel(0) 
        file_handler.close() 
        ftp.quit()
    
    
    

