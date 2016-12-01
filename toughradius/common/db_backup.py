#!/usr/bin/env python
#coding:utf-8
from sqlalchemy import *
from toughradius.common.dbengine import get_engine
import json,os,gzip

class DBBackup:

    def __init__(self, sqla_metadata, excludes=[],batchsize=49,**kwargs):
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

    def restoredbv1(self,restorefs):
        if not os.path.exists(restorefs):
            print 'backup file not exists'
            return

        table_defines = {
            'slc_node' : 'tr_node',
            'slc_operator' : 'tr_operator',
            'slc_rad_operate_log' : 'tr_operate_log',
            'slc_rad_accept_log' : 'tr_accept_log',
            'slc_rad_bas' : 'tr_bas',
            'slc_member' : 'tr_customer',
            'slc_member_order' : 'tr_customer_order',
            'slc_rad_account' : 'tr_account',
            'slc_rad_product' : 'tr_product',
            'slc_rad_product_attr' : 'tr_product_attr',
        }
        with self.dbengine.begin() as db:
            with gzip.open(restorefs,'rb') as rfs:
                for line in rfs:
                    flag = False
                    for t in table_defines:
                        if t not in line:
                            flag = True 
                            break

                    if not flag:
                        continue
                    line = line.replace('member_id','customer_id')
                    line = line.replace('member_name','customer_name')
                    line = line.replace('member_desc','customer_desc')
                    line = line.replace('vlan_id','vlan_id1')
                    line = line.replace('vlan_id12','vlan_id2')
                    try:
                        obj = json.loads(line)
                        ctable = table_defines.get(obj['table'])
                        if not ctable:
                            continue
                        print "delete from %s"%ctable
                        db.execute("delete from %s"%ctable)
                        print self.metadata.tables[ctable].insert()
                        objs =  obj['data']
                        if len(objs) < 49:
                            if objs:
                                print 'insert %s data' % len(objs)
                                db.execute(self.metadata.tables[ctable].insert().values(objs))
                        else:
                            while len(objs) > 0:
                                _tmp_pbjs = objs[:49]
                                objs = objs[49:]
                                print 'insert %s data' % len(_tmp_pbjs)
                                db.execute(self.metadata.tables[ctable].insert().values(_tmp_pbjs))
                            
                        # db.execute("commit;")
                    except:
                        print 'error data %s ...'%line[:128] 
                        import traceback
                        traceback.print_exc()


if __name__ == '__main__':
    pass







