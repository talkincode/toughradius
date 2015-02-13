#!/usr/bin/env python
#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
import argparse,ConfigParser
from datetime import datetime
from ftplib import FTP

def backup(**kwargs):
    import sh
    bakdir = "/var/toughradius/databak"
    if not os.path.exists(bakdir):
        os.mkdir(bakdir)
    now = datetime.now()
    dbname = kwargs.pop('dbname','toughradius')
    ftphost = kwargs.pop('ftphost','127.0.0.1')
    ftpport = kwargs.pop('ftpport',21)
    ftpuser = kwargs.pop('ftpuser','')
    ftppwd = kwargs.pop('ftppwd','')
    backfile = '%s/%s-backup-%s.gz'%(bakdir,dbname,now.strftime( "%Y%m%d"))
    
    sh.gzip(sh.mysqldump(u='root',B=dbname,S="/var/toughradius/mysql/mysql.sock"),'-cf',_out=backfile)

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
    

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-c','--conf', type=str,default='./radiusd.conf',dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])    
    # read config file
    config = ConfigParser.ConfigParser()
    config.read(args.conf)
    
    backup(
        dbname=config.get('database','db'),
        ftphost=config.get('database','ftphost'),
        ftpport=config.get('database','ftpport'),
        ftpuser=config.get('database','ftpuser'),
        ftppwd=config.get('database','ftppwd')                        
    )
    

