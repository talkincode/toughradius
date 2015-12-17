#!/usr/bin/env python
#coding:utf-8
import json
import os
import gzip

from sqlalchemy import *

from toughradius.tools.dbengine import get_engine
from toughradius.console import models

excludes = ['slc_rad_ticket','slc_rad_billing','slc_rad_online']

def dumpdb(config,dumpfs):
    _dir = os.path.split(dumpfs)[0]
    if not os.path.exists(_dir):
        os.makedirs(_dir)

    engine = get_engine(config)
    db = engine.connect()
    metadata = models.get_metadata(engine)
    with gzip.open(dumpfs,'wb') as dumpfs:
        for _name,_table in metadata.tables.items():
            if _name in excludes:
                continue
            rows = db.execute(select([_table])).fetchall()
            obj = dict(table=_name,data=[dict(r.items()) for r in rows])
            dumpfs.write(json.dumps(obj,ensure_ascii=False).encode('utf-8'))
            dumpfs.write('\n')
    db.close()


def restoredb(config,restorefs):
    if not os.path.exists(restorefs):
        print 'backup file not exists'
    else:
        engine = get_engine(config)
        db = engine.connect()
        metadata = models.get_metadata(engine)
        with gzip.open(restorefs,'rb') as rfs:
            for line in rfs:
                try:
                    obj = json.loads(line)
                    print "delete from %s"%obj['table']
                    db.execute("delete from %s"%obj['table'])
                    print 'insert datas into %s'%obj['table']
                    objs =  obj['data']
                    if len(objs) < 500:
                        if objs:db.execute(metadata.tables[obj['table'] ].insert().values(objs))
                    else:
                        while len(objs) > 0:
                            _tmp_pbjs = objs[:500]
                            objs = objs[500:]
                            db.execute(metadata.tables[obj['table'] ].insert().values(_tmp_pbjs))
                        
                    # db.execute("commit;")
                except:
                    print 'error data %s ...'%line[:128] 
                    import traceback
                    traceback.print_exc()
        db.close()
