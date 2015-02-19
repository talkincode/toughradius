#!/usr/bin/env python
#coding:utf-8

import session

routers = [
    '/',
    '/cache/clean',
    '/login',
    '/param',
    '/node',
    '/node/add',
    '/node/update?node_id=0',
    '/bas',
    '/bas/add',
    '/bas/update?bas_id=0',
    '/opr',
    '/opr/add',
    '/roster',
    '/roster/add',
    '/bus/member',
    '/bus/member/open',
    '/bus/member/import',
    '/bus/acceptlog',
    '/bus/billing',
    '/bus/orders',
    '/card/list',
    '/card/create',
    '/ops/user',
    '/ops/user/trace',
    '/ops/online',
    '/ops/ticket',
    '/ops/opslog',
    '/ops/online/stat',
    '/ops/flow/stat',
    '/product',
    '/product/add',
    '/product/attr/add'
]


def test_routers():
    req = session.login()
    for r in routers:
        _path = session.sub_path(r)
        print _path
        r = req.get(_path)
        assert r.status_code == 200

if __name__ == '__main__':
    test_routers()
