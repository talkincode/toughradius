#!/usr/bin/env python
#coding:utf-8

import session

member = dict(
    node_id = 1,realname = u'测试用户',
    member_name = u'tester',member_password = '888888',
    idcard = '432501090897973',mobile = '13882880172',
    address = u'测试地址',account_number = 'test01',
    password = '888888',ip_address = '192.168.0.1',
    product_id = 0,months = 1,
    fee_value = 30,expire_date = '2015-03-20',
    status = 1
)

def test_post_member_100():
    req = session.login()    
    r0 = req.get(session.sub_path(u"/test/pid?name=预付费包月30元"))
    assert r0.status_code == 200
    pid0 = r0.json()['pid']
    
    for i in range(100):
        memberi = member.copy()
        memberi['member_name'] = 'ppmuser%s'%(i+1)
        memberi['realname'] = u"测试包月用户%s"%(i+1)
        memberi['account_number'] = 'ppm%s'%(i+1)
        memberi['ip_address'] = '192.168.1.%s'%(i+1)
        memberi['product_id'] = pid0
        r = req.post(session.sub_path("/bus/member/open"),data=memberi)
        assert r.status_code == 200
        assert '<span class="wrong">' not in  r.text
        
    

        
def test_post_member():
    req = session.login()    
    r0 = req.get(session.sub_path(u"/test/pid?name=预付费包月30元"))
    assert r0.status_code == 200
    r1 = req.get(session.sub_path(u"/test/pid?name=预付费时长每小时2元"))
    assert r1.status_code == 200
    r2 = req.get(session.sub_path(u"/test/pid?name=买断包月12个月500元"))
    assert r2.status_code == 200
    r3 = req.get(session.sub_path(u"/test/pid?name=买断时长100元50小时"))
    assert r3.status_code == 200
    r4 = req.get(session.sub_path(u"/test/pid?name=预付费流量每MB0.05元"))
    assert r4.status_code == 200
    r5 = req.get(session.sub_path(u"/test/pid?name=买断流量5元100MB"))
    assert r5.status_code == 200
    
    # 快速开新户
    pid0 = r0.json()['pid']
    member['product_id'] = pid0
    r = req.post(session.sub_path("/bus/member/open"),data=member)
    assert r.status_code == 200
    assert '<span class="wrong">' not in  r.text
    # 获取memberid
    r = req.get(session.sub_path(u"/test/mid?name=tester"))
    assert r.status_code == 200
    mid = r.json()['mid']
    #新增账号
    for _r in (r0,r1,r2,r3,r4,r5):
        pid = _r.json()['pid']
        account = dict(
            node_id = 1,
            member_id = mid,
            realname = u"测试用户",
            account_number = "test00%s"%pid,
            password = "888888",
            ip_address = '',
            address = 'test address',
            product_id = pid,
            months = 12,
            fee_value = '100.00',
            expire_date =  "2500-12-30",
            status = 1
        )
        rr = req.post(session.sub_path("/bus/account/open"),data=account)
        assert rr.status_code == 200
        assert '<span class="wrong">' not in  r.text
        