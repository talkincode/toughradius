#!/usr/bin/env python
#coding:utf-8
from sqlalchemy import *
from toughradius.common.dbengine import get_engine
import json,os,gzip

class DBBackup:

    def __init__(self, sqla_metadata, excludes=[],batchsize=32,**kwargs):
        self.metadata = sqla_metadata
        self.excludes = excludes
        self.dbengine = self.metadata.bind
        self.batchsize = batchsize

    def dumpdb(self, dumpfile):
        _dir = os.path.split(dumpfile)[0]
        if not os.path.exists(_dir):
            os.makedirs(_dir)

        with gzip.open(dumpfile, 'wb') as dumpfs:
            tables = {_name:_table for _name, _table in self.metadata.tables.items() if _name not in self.excludes}
            table_headers = ('table_names', tables.keys())
            dumpfs.write(json.dumps(table_headers, ensure_ascii=False).encode('utf-8'))
            dumpfs.write('\n')
            for _name,_table in tables.iteritems():
                with self.dbengine.begin() as db:
                    rows = db.execute(select([_table]))
                    for row in rows:
                        obj = (_name, dict(row.items()))
                        dumpfs.write(json.dumps(obj,ensure_ascii=False).encode('utf-8'))
                        dumpfs.write('\n')



    def restoredb(self,restorefs):
        if not os.path.exists(restorefs):
            print 'backup file not exists'
            return
        
        with gzip.open(restorefs,'rb') as rfs:
            cache_datas = {}
            for line in rfs:
                try:
                    with self.dbengine.begin() as db:
                        tabname, rdata = json.loads(line)
                        if tabname == 'table_names' and rdata:
                            for table_name in rdata:
                                print "clean table %s" % table_name
                                db.execute("delete from %s;" % table_name)
                            continue

                        cache_datas.setdefault(tabname, []).append(rdata)

                        if len(cache_datas[tabname]) >= self.batchsize:
                            print 'batch insert datas<%s> into %s' % (len(cache_datas[tabname]), tabname)
                            db.execute(self.metadata.tables[tabname].insert().values(cache_datas[tabname]))
                            del cache_datas[tabname]
                except:
                    print 'error data %s ...'% line
                    raise

            print "insert last data"
            for tname, tdata in cache_datas.iteritems():
                try:
                    print 'batch insert datas<%s> into %s' % (len(tdata), tname)
                    with self.dbengine.begin() as db:
                        db.execute(self.metadata.tables[tname].insert().values(tdata))
                except:
                    print 'error data %s ...' % tdata
                    raise

            cache_datas.clear()

if __name__ == '__main__':
    pass







