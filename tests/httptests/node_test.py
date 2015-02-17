#!/usr/bin/env python
#coding:utf-8

import session

def test_post_node():
    req = session.login()
    r = req.post(session.sub_path("/node/add"),dict(node_name="testnode",node_desc=u"测试区域2"))
    assert r.status_code == 200

if __name__ == '__main__':
    test_post_node()