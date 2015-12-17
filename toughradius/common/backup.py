#!/usr/bin/env python
#coding:utf-8
from sqlalchemy import *
from toughradius.common.dbengine import get_engine
from toughradius.console import models
import json,os,gzip

excludes = ['tr_online']

def dumpdb(config,dumpfs):
    _dir = os.path.split(dumpfs)[0]
    if not os.path.exists(_dir):
        os.makedirs(_dir)

    engine = get_engine(config)
    db = engine.connect()
    metadata = models.get_metadata(engine)
    with gzip.open(dumpfs, 'wb') as dumpfs:

        table_names = [_name for _name, _ in metadata.tables.items()]
        table_headers = ('table_names', table_names)
        dumpfs.write(json.dumps(table_headers, ensure_ascii=False).encode('utf-8'))
        dumpfs.write('\n')

        for _name,_table in metadata.tables.items():
            if _name in excludes:
                continue
            rows = db.execute(select([_table]))
            for rows in rows:
                obj = (_name, dict(rows.items()))
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
            cache_datas = {}
            for line in rfs:
                try:
                    tabname, rdata = json.loads(line)

                    if tabname == 'table_names' and rdata:
                        for table_name in rdata:
                            print "clean table %s" % table_name
                            db.execute("delete from %s;" % table_name)
                        continue

                    if tabname not in cache_datas:
                        cache_datas[tabname] = [rdata]
                    else:
                        cache_datas[tabname].append(rdata)

                    if tabname in cache_datas and len(cache_datas[tabname]) >= 500:
                        print 'insert datas<%s> into %s' % (len(cache_datas[tabname]), tabname)
                        db.execute(metadata.tables[tabname].insert().values(cache_datas[tabname]))
                        del cache_datas[tabname]

                except:
                    print 'error data %s ...'% line
                    import traceback
                    traceback.print_exc()

            print "insert last data"
            for tname, tdata in cache_datas.iteritems():
                try:
                    print 'insert datas<%s> into %s' % (len(tdata), tname)
                    db.execute(metadata.tables[tname].insert().values(tdata))
                except:
                    print 'error data %s ...' % tdata
                    import traceback
                    traceback.print_exc()

            cache_datas.clear()

        db.close()


