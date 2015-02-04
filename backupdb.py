#!/usr/bin/env python
#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
import argparse,json
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

    if '127.0.0.1' not in ftphost:
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
    parser.add_argument('-c','--conf', type=str,default='./config.json',dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])    
    config=json.loads(open(args.conf,'rb').read())
    
    backup(**dict(
        dbname= config['database']['db'],
        ftphost= config['backup']['ftphost'],
        ftpuser= config['backup']['ftpuser'],
        ftppwd= config['backup']['ftppwd']                                       
    ))
    

