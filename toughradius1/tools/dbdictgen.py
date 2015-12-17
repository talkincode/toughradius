#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
from toughradius.console import models

def print_header():
    print "%s  %s  %s  %s"%("="*21,"="*16,"="*16,"="*36)
    print "%s%s%s%s"%("属性".ljust(26,' '),"类型（长度）".ljust(25,' '),"可否为空".ljust(23,' '),'描述'.ljust(36,' '))
    print "%s  %s  %s  %s"%("="*21,"="*16,"="*16,"="*36)

def print_end():
    print "%s  %s  %s  %s"%("="*21,"="*16,"="*16,"="*36)
    print '\n.. end_table\n'

def print_model(tmdl):
    print '.. _%s_label:\n'%tmdl.__tablename__
    print tmdl.__tablename__
    print '-'*36,"\n"
    if tmdl.__doc__ :
        print tmdl.__doc__
        print

    pk = ",".join( c.name for c in tmdl.__table__.primary_key.columns)
    print ".. start_table %s;%s"%(tmdl.__tablename__,pk),"\n"
    print_header()
    for c in  tmdl.__table__.columns:
        # print c.name,c.type,c.nullable
        _name = str(c.name).ljust(21," ")
        _type = str(c.type).ljust(16," ")
        _null = str(c.nullable).ljust(16," ")
        _doc = str((c.doc or '').encode("utf-8")).ljust(26,' ')
        print "%s  %s  %s  %s"%(_name,_type,_null,_doc)
    print_end()

mdls = [
    models.SlcNode,
    models.SlcOperator,
    models.SlcOperatorRule,
    models.SlcParam,
    models.SlcRadBas,
    models.SlcRadRoster,
    models.SlcMember,
    models.SlcMemberOrder,
    models.SlcRadAccount,
    models.SlcRadAccountAttr,
    models.SlcRadProduct,
    models.SlcRadProductAttr,
    models.SlcRadBilling,
    models.SlcRadTicket,
    models.SlcRadOnline,
    models.SlcRadOnlineStat,
    models.SlcRadFlowStat,
    models.SlcRadAcceptLog,
    models.SlcRadOperateLog,
    models.SlcRechargerCard,
    models.SlcRechargeLog
]


def main():
    print "ToughRADIUS数据字典\n%s\n\n"%('='*36)
    for mdl in mdls:
        print_model(mdl)
        print

