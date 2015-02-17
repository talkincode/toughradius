#!/usr/bin/env python
#coding:utf-8

import session

def test_post_bas():
    bas = dict(
            ip_addr="192.168.88.1",bas_name="stdbas",
            bas_secret="123456",vendor_id='14988',
            coa_port='3799',time_type='0'
        )
    print 'post bas',bas
    req = session.login()
    r = req.post(session.sub_path("/bas/add"),bas)
    assert r.status_code == 200

if __name__ == '__main__':
    test_post_bas()