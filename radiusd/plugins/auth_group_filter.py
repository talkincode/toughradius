#!/usr/bin/env python
#coding=utf-8
from plugins import error_auth
from store import store

def process(req=None,resp=None,user=None):
    """执行用户组策略校验，检查MAC与VLANID绑定，并发数限制 """
    group = store.get_group(user['group_id'])

    if not group:
        return resp

    if group['bind_mac']:
        if user['mac_addr'] and get_mac_addr() not in user['mac_addr']:
            return error_auth(resp,"macaddr not match")
        if not user['mac_addr']:
            user['mac_addr'] = get_mac_addr()
            store.update_user_mac(user['account_number'],get_mac_addr())

    if group['bind_vlan']:
        vlan_id,vlan_id2 = req.get_vlanids()
        #update user vlan_bind
        if vlan_id and user['vlan_id']:
            if vlan_id != user['vlan_id']:
                return error_auth(resp,"vlan_id bind not match")
        elif vlan_id and not user['vlan_id']:
            user['vlan_id'] = vlan_id
            store.update_user_vlan_id(user['account_number'],vlan_id)

        if vlan_id2 and user['vlan_id2']:
            if vlan_id2 != user['vlan_id2']:
                return error_auth(resp,"vlan_id2 bind not match")
        elif vlan_id2 and not user['vlan_id2']:
            user['vlan_id2'] = vlan_id2
            store.update_user_vlan_id2(user['account_number'],vlan_id2)

    return resp