#coding:utf-8
from collections import namedtuple

class Constant(dict):
    def __getattr__(self, key):
        try:
            return self[key]
        except KeyError as k:
            raise AttributeError(k)
    def __setattr__(self, key, value):
        pass
    def __delattr__(self, key):
        pass
    def __repr__(self):
        return '<Constant ' + dict.__repr__(self) + '>'

# 0:nasid 1:nasip
NasHKey = "toughradius:nas_hkey:{0}:{1}".format
NasSetKey = "toughradius:nas_set"

UserSetKey = "toughradius:user_set_key"
# 0:username
UserHKey = "toughradius:user_hkey:{0}".format

# 0:username
UserRadAttrsHKey = "toughradius:user_radattrs_hkey:{0}".format

# 0:nasid 1:sessionid
OnlineHKey = "toughradius:online_hkey:{0}:{1}".format
OnlineSetKey = "toughradius:online_set_key"
# 0:username
UserOnlineSetKey = "toughradius:user_online_set_key:{0}".format
# 0:nasid 1:nasip
NasOnlineSetKey = "toughradius:nas_online_set_key:{0}:{1}".format


NameAttr = namedtuple('NameAttr', 'name desc')
ValueAttr = namedtuple('ValueAttr', 'value desc')

UserStates = Constant(
    normal=ValueAttr(value=1,desc="user normal status"),
    pause=ValueAttr(value=0,desc="user pause status")
)

UserAttrs = Constant(
    status=NameAttr(name='status',desc='user status 0/1'),
    username=NameAttr(name='username',desc='user name'),
    password=NameAttr(name='password',desc='user password string'),
    input_rate=NameAttr(name='input_rate',desc='user input limit (bps)'),
    output_rate=NameAttr(name='output_rate',desc='user output limit (bps)'),
    rate_code=NameAttr(name='rate_code',desc='user limit code'),
    bill_type=NameAttr(name='bill_type',desc='user billing type day/second'),
    bind_mac=NameAttr(name='bind_mac',desc='user check mac bind 0/1'),
    bind_vlan=NameAttr(name='bind_vlan',desc='user check vlan bind 0/1'),
    bind_nas=NameAttr(name='bind_nas',desc='user check nas bind 0/1'),
    nas=NameAttr(name='nas',desc='nas id or ipaddr'),
    mac_addr=NameAttr(name='mac_addr',desc='user mac addr'),
    vlanid1=NameAttr(name='vlanid1',desc='user vlanid1'),
    vlanid2=NameAttr(name='vlanid2',desc='user vlanid2'),
    time_amount=NameAttr(name='time_amount',desc='user time length (seconds)'),
    flow_amount=NameAttr(name='flow_amount',desc='user flow length (kb)'),
    expire_date=NameAttr(name='expire_date',desc='user expire date format:yyyy-mm-dd'),
    expire_time=NameAttr(name='expire_time',desc='user expire time format:hh:mm:ss'),
    online_limit=NameAttr(name='online_limit',desc='user  online limit'),
    bypass_pwd=NameAttr(name='bypass_pwd',desc='user ignore password check')
)

NasStates = Constant(
    normal=ValueAttr(value=1,desc="nas normal status"),
    pause=ValueAttr(value=0,desc="nas pause status")
)

NasAttrs = Constant(
    status=NameAttr(name='status',desc='nas status 0/1'),
    nasid=NameAttr(name='nasid',desc='nas id'),
    name=NameAttr(name='name',desc='nas name'),
    vendor=NameAttr(name='vendor',desc='nas vendor id'),
    ipaddr=NameAttr(name='ipaddr',desc='nas ip address'),
    secret=NameAttr(name='secret',desc='nas secret'),
    coaport=NameAttr(name='coaport',desc='nas coa service port')
)





