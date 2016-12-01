#!/usr/bin/env python
#coding:utf-8
import json

class make_db:
    def __init__(self, mkdb):
        self.conn = mkdb()

    def __enter__(self):
        return self.conn   

    def __exit__(self, exc_type, exc_value, exc_tb):
        self.conn.close()

def serial_model(mdl):
    if not mdl:return
    if not hasattr(mdl,'__table__'):return
    data = {}
    for c in mdl.__table__.columns:
        data[c.name] = getattr(mdl, c.name)
    return json.dumps(data,ensure_ascii=False)