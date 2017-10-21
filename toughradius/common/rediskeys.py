#coding:utf-8

NasIdHKey = "toughradius:nasid_hkey:{nasid}".format
NasIpHKey = "toughradius:nasip_hkey:{nasip}".format
NasSetKey = "toughradius:nas_set"

UserSetKey = "toughradius:user_set_key"
UserHKey = "toughradius:user_hkey:{username}".format
UserRadAttrsHKey = "toughradius:user_radattrs_hkey:{username}".format

OnlineHKey = "toughradius:online_hkey:{nasid}:{sessionid}".format
OnlineSetKey = "toughradius:online_set_key"
UserOnlineSetKey = "toughradius:user_online_set_key:{username}".format
NasOnlineSetKey = "toughradius:nas_online_set_key:{nasid}".format