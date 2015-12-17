#!/usr/bin/env python
#coding=utf-8
from toughradius.radiusd.plugins import error_auth

def process(req=None,resp=None,user=None,radiusd=None,**kwargs):
    """check mac bind & vlan bind"""
    store = radiusd.store
    macaddr = req.get_mac_addr()
    if store.is_white_roster(macaddr):
        return resp
        
    if macaddr and  user['mac_addr']:
        if user['bind_mac'] == 1 and macaddr not in user['mac_addr']:
            return error_auth(resp,"macaddr bind not match")
    elif macaddr and not user['mac_addr'] :
        store.update_user_mac(user['account_number'], macaddr)

    vlan_id,vlan_id2 = req.get_vlanids()
    print vlan_id,vlan_id2
    if vlan_id and user['vlan_id']:
        if user['bind_vlan'] == 1 and vlan_id != user['vlan_id']:
            return error_auth(resp,"vlan_id bind not match")
    elif vlan_id and not user['vlan_id']:
        user['vlan_id'] = vlan_id
        store.update_user_vlan_id(user['account_number'],vlan_id)

    if vlan_id2 and user['vlan_id2']:
        if user['bind_vlan'] == 1 and vlan_id2 != user['vlan_id2']:
            return error_auth(resp,"vlan_id2 bind not match")
    elif vlan_id2 and not user['vlan_id2']:
        user['vlan_id2'] = vlan_id2
        store.update_user_vlan_id2(user['account_number'],vlan_id2)

    return resp