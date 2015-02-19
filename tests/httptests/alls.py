#!/usr/bin/env python
#coding:utf-8

import bas_test
import node_test 
import product_test
import member_test
import admin_router_test

if __name__ == '__main__':
    node_test.test_post_node()
    bas_test.test_post_bas()
    product_test.test_post_product()
    member_test.test_post_member()
    admin_router_test.test_routers()
    # member_test.test_post_member_100()