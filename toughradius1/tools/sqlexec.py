#!/usr/bin/env python
#coding:utf-8
import os
from toughradius.tools.dbengine import get_engine
from sqlalchemy.sql import text as _sql
from toughradius.tools.shell import shell as sh

def _decode(src):
    try:
        return unicode(src,'utf-8')
    except:
        return unicode(src)

def print_result(rs):
    n = 0
    rstr = []
    for r in rs:
        if n == 100:
            return
        try:
            _r = ', '.join([ '%s=%s'%(k,_decode(r[k])) for k in r.keys()])
            rstr.append(_r)
        except Exception as err:
            sh.warn('print result error: %s'%str(err))
    sh.info('Result:\n\n%s\n'%('\n'.join(rstr)))        
    
def execute_sqls(config,sqlstr):
    sh.info('exec sql >> %s'%sqlstr)
    results = []
    with get_engine(config).begin() as conn:
        try:
            results = conn.execute(_sql(sqlstr))
        except Exception as err:
            return sh.err('exec sql error: %s'%str(err))
    sh.info('exec sql done')
    print_result(results)
    

def execute_sqlf(config,sqlfile):
    sh.info('exec sql file >> %s'%sqlfile)
    if not os.path.exists(sqlfile):
        return sh.warn('sqlfile not exists')
    conn = get_engine(config).raw_connection()
    try:
        for line in open(sqlfile):
            sh.info(line)
            conn.execute(line)
    except Exception as err:
        return sh.err('exec sql file error: %s'%str(err))
    finally:
        conn.close()
    sh.info('exec sql file done')

    
    
    