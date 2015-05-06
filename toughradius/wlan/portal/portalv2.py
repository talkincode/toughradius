#!/usr/bin/env python
#coding:utf-8
import struct
import random
import hashlib
import pktutils
import binascii
import six
import copy
import itertools
import socket

md5_constructor = hashlib.md5

random_generator = random.SystemRandom()

__CurrentSN = random_generator.randrange(1, 32767)

"""
REQ_CHALLENGE	0x01	Portal server--> BAS	表示此报文是Portal Server向BAS发送的 Challenge请求报文	必须
ACK_CHALLENGE	0x02	BAS --> Portal server	表示此报文是BAS对Portal Server请求Challenge报文的响应报文	必须
REQ_AUTH	0x03	Portal server --> BAS	表示此报文是Portal Server向BAS发送的请求认证报文	必须
ACK_AUTH	0x04	BAS --> Portal server	表示此报文是BAS对Portal Server请求认证报文的响应报文	必须
REQ_LOGOUT	0x05	Portal server --> BAS	表示此报文是Portal  Server向BAS发送的下线请求报文	必须
ACK_LOGOUT	0x06	BAS --> Portal server	表示此报文是BAS对Portal Server下线请求的响应报文	必须
AFF_ACK_AUTH	0x07	Portal server --> BAS	表示此报文是Portal Server收到认证成功响应报文后向BAS发送的确认报文	建议
NTF_LOGOUT	0x08	BAS --> Portal server	表示此报文是BAS发送给Portal Server，用户被强制下线的通知报文	必须
REQ_INFO	0x09	Portal server --> BAS	信息询问报文	必须
ACK_INFO 	0x0a	BAS --> Portal server	信息询问的应答报文	必须
NTF_USERDISCOVER	0x0b	Portal server --> BAS	Portal Server向BAS发送的发现新用户要求上线的通知报文	建议
NTF_USERIPCHANGE 	0x0c	BAS --> Portal server	BAS向Portal Server发送的通知更改某个用户IP地址的通知报文	必须
AFF_NTF_USERIPCHAN	0x0d	Portal server --> BAS	PortalServer通知BAS对用户表项的IP切换已成功	必须
ACK_NTF_LOGOUT	0x0e	Portal server --> BAS	PortalServer通知BAS用户强制下线成功，BAS通过NTF_LOGOUT报文通知Portal Server用户下线后，Portal Server回应BAS设备用户下线完成的回应报文。如果Portal Server收到了BAS的用户下线请求，必须回应ACK_NTF_LOGOUT，以通知BAS服务器，无论用户是否在线。同时，Portal Server必须确保用户下线处理成功。	必须
NTF_HEARTBEAT	0x0f	Portal server --> BAS	逃生心跳报文，PortalServer周期性的向BAS发送该报文，以表明PortalServer可以正常提供服务。BAS如果连续多次没有接收到该报文，说明PortalServer已经停止服务，BAS即切换为逃生状态，此时不再强制用户认证，允许用户的报文直接通过。该报文没有回应报文。	必须
NTF_USER_HEARTBEAT	0x10	Portal server --> BAS	用户心跳报文，PortalServer周期性的向BAS发送该报文，以表明该用户仍然在线，BAS如果连续多次没有接收到含有该用户IP的报文，说明该用户已经断线，BAS将向RADIUS服务器发送下线报文，将用户下线。用户心跳报文中包含了多个用户的IP地址。	必须
ACK_NTF_USER_HEARTBEAT	0x11	BAS --> Portal server	用户心跳回应报文，BAS接收到PortalServer的用户心跳报文后，会遍历这些用户IP地址，并将已经下线的用户IP地址放入回应报文中。PortalServer收到回应报文后，将用户下线。如果用户心跳报文中的所有用户都在线，则BAS将不发送回应报文。	必须
NTF_CHALLENGE	0x12	BAS --> Portal server	表示此报文是 BAS 向Portal Server 发送的Challenge请求报文，主要适用于EAP_TLS认证。	建议
NTF_USER_NOTIFY	0x13	BAS --> Portal server	用户消息通知报文。在Pap/Chap认证方式下，计费回应报文中Radius服务器需要向用户下发一些消息，例如帐号余额等信息。	建议
AFF_NTF_USER_NOTIFY	0x14	Portal server --> BAS	PortalServer通知BAS消息已收到	建议
"""

REQ_CHALLENGE = 0x01
ACK_CHALLENGE = 0x02
REQ_AUTH = 0x03
ACK_AUTH = 0x04
REQ_LOGOUT = 0x05
ACK_LOGOUT = 0x06
AFF_ACK_AUTH = 0x07
NTF_LOGOUT = 0x08
REQ_INFO = 0x09
ACK_INFO = 0x0a
NTF_USERDISCOVER = 0x0b
NTF_USERIPCHANGE = 0x0c
AFF_NTF_USERIPCHAN = 0x0d
ACK_NTF_LOGOUT = 0x0e
NTF_HEARTBEAT = 0x0f
NTF_USER_HEARTBEAT = 0x10
ACK_NTF_USER_HEARTBEAT = 0x11
NTF_CHALLENGE = 0x12
NTF_USER_NOTIFY = 0x13
AFF_NTF_USER_NOTIFY = 0x14
 
AUTH_CHAP = 0x00
AUTH_PAP = 0x01

def CurrentSN():
    global __CurrentSN
    __CurrentSN = (__CurrentSN + 1) % 32767
    return __CurrentSN
    
def hexlify(attr_type,value):
    if attr_type in (0x03,0x04,0x0a):
        return binascii.hexlify(value)
    else:
        return value
        

AckChallengeErrs = {
    1 : u"请求Challenge被拒绝",
    2 : u"链接已建立",
    3 : u"有一个用户正在认证过程中，请稍后再试",
    4 : u"用户请求Challenge失败"
}

AckAuthErrs = {
    1 : u"用户认证请求被拒绝",
    2 : u"链接已建立",
    3 : u"有一个用户正在认证过程中，请稍后再试",
    4 : u"用户认证失败"
}

AckLogoutErrs = {
    1 : u'用户下线被拒绝',
    2 : u'用户下线失败（发生错误）',
    3 : u'用户已经下线'
}

AckInfoErrs = {
    1 : '不支持信息查询功能或者处理失败',
    2 : '消息处理失败，由于某种不可知原因'
}

PKT_TYPES = {
    REQ_CHALLENGE   : 'REQ_CHALLENGE',
    ACK_CHALLENGE   : 'ACK_CHALLENGE',
    REQ_AUTH   : 'REQ_AUTH',
    ACK_AUTH   : 'ACK_AUTH',
    REQ_LOGOUT   : 'REQ_LOGOUT',
    ACK_LOGOUT   : 'ACK_LOGOUT',
    AFF_ACK_AUTH   : 'AFF_ACK_AUTH',
    NTF_LOGOUT   : 'NTF_LOGOUT',
    REQ_INFO   : 'REQ_INFO',
    ACK_INFO   : 'ACK_INFO',
    NTF_USERDISCOVER   : 'NTF_USERDISCOVER',
    NTF_USERIPCHANGE   : 'NTF_USERIPCHANGE',
    AFF_NTF_USERIPCHAN   : 'AFF_NTF_USERIPCHAN',
    ACK_NTF_LOGOUT   : 'ACK_NTF_LOGOUT',
    NTF_HEARTBEAT   : 'NTF_HEARTBEAT',
    NTF_USER_HEARTBEAT   : 'NTF_USER_HEARTBEAT',
    ACK_NTF_USER_HEARTBEAT   : 'ACK_NTF_USER_HEARTBEAT',
    NTF_CHALLENGE   : 'NTF_CHALLENGE',
    NTF_USER_NOTIFY   : 'NTF_USER_NOTIFY',
    AFF_NTF_USER_NOTIFY   : 'AFF_NTF_USER_NOTIFY'
}


class Error(Exception): pass
class UnpackError(Error): pass
class NeedData(UnpackError): pass
class PackError(Error): pass

class PortalV2(object):
    
    __hdr__ = (
        ('ver', 'B', 0x02),
        ('type', 'B', 0),
        ('isChap', 'B', 0),
        ('rsv', 'B', 0),
        ('serialNo', 'H', 0),
        ('reqId', 'H', 0),
        ('userIp', '4s', 0),
        ('userPort', 'H', 0),
        ('errCode', 'B', 0),
        ('attrNum', 'B', 0),
        ('auth', '16s', '')
    )
    attrs = []
    
    def __init__(self,**attributes):
        # init hdr
        self.__hdr_fmt__ = '>'+''.join([ x[1] for x in self.__hdr__ ])
        self.__hdr_fields__ = [ x[0] for x in self.__hdr__ ]
        self.__hdr_len__ = struct.calcsize(self.__hdr_fmt__)
        self.__hdr_defaults__ = dict(zip(self.__hdr_fields__, [ x[2] for x in self.__hdr__ ]))
        
        if 'secret' in attributes:
            self.secret = attributes.pop('secret')
            
        if 'source' in attributes:
            self.source = attributes.pop('source')

        if 'packet' in attributes:
            try:
                self.unpack(attributes.pop('packet'))
            except struct.error:
                if len(args[0]) < self.__hdr_len__:
                    raise NeedData
                raise UnpackError('invalid %s: %r' %(self.__class__.__name__, args[0]))
        else:
            for k in self.__hdr_fields__:
                setattr(self, k, copy.copy(self.__hdr_defaults__[k]))
            for k, v in attributes.iteritems():
                setattr(self, k, v)

        for (key, value) in attributes.items():
            if key in self.__hdr_fields__:
                setattr(self,key,value)
    @property          
    def sid(self):
        return "{0}_{1}".format(self.serialNo,self.userIp)

    def __len__(self):
        return self.__hdr_len__ + sum([ 2 + len(o[1]) for o in self.attrs ]) 
        
    def __str__(self):
        return self.pack_hdr() + self.pack_attrs()
        
    def __repr__(self):
        l = [ '%s=%r' % (k, getattr(self, k)) for k in self.__hdr_defaults__ ]
        l.append("attrs=%s"%(self.attrs))        
        return '%s <%s> (%s)' % (self.__class__.__name__,PKT_TYPES.get(self.type), ', '.join(l))
        
    def auth_packet(self):
        _auth = md5_constructor(str(self)+self.secret).digest()
        self.auth = _auth

    def check_resp_auth(self,req_auth):
        resp_auth = self.auth
        self.auth = req_auth
        _auth = md5_constructor(str(self)+self.secret).digest()
        self.auth = resp_auth
        return resp_auth == _auth
            
    def pack_hdr(self):
        """Return packed header string."""
        try:
            params = [ getattr(self, k) for k in self.__hdr_fields__ ]
            return struct.pack(self.__hdr_fmt__,*params)
        except Exception as e:
            raise PackError(e)
           
    def pack_attrs(self):
        if not self.attrs:
            return ''
        l = []
        for t, data in self.attrs:
            l.append('%s%s%s' % (chr(t), chr(len(data)+2), data))
        return ''.join(l)
        
    def pack(self):
        """Return packed header + self.data string."""
        return str(self)
    
    def unpack(self, buf):
        #parse header
        for k, v in itertools.izip(self.__hdr_fields__,
            struct.unpack(self.__hdr_fmt__, buf[:self.__hdr_len__])):
            setattr(self, k, v)
        # parse attrs    
        buf = buf[self.__hdr_len__:]
        self.attrs = []
        _count = 0
        while buf:
            if _count == self.attrNum:
                return
            t = ord(buf[0])
            l = ord(buf[1])
            d, buf = buf[2:l], buf[l:]
            self.attrs.append((t, d))
            _count += 1

            

    def get_user_name(self):
        """
        长度: <=16(可变)
        描述: 该属性表示要认证的用户的名字，必须出现在REQ_AUTH（03）报文中。
        该属性表明了待验证的用户的密码，在传输过程中是加密的。当用户采用PAP方式认证时，必须出现在REQ_AUTH（03）报文中。
        """
        for attr in self.attrs:
            if attr[0] == 0x01:
                return pktutils.DecodeString(attr[1])
              
    def get_password(self):
        """
        长度: 16(固定)
        描述: 该属性表明了待验证的用户的密码，在传输过程中是加密的。当用户采用PAP方式认证时，必须出现在REQ_AUTH（03）报文中。
        此属性值的长度要求少于32字节。
        """
        for attr in self.attrs:
            if attr[0] == 0x02:
                return pktutils.DecodeString(attr[1])

                
    def get_challenge(self):
        """
        长度: 16(固定)
        描述: 表示设备传送给CHAP认证的用户的chap challenge。
        它只能用在chap方式认证的REQ_AUTH(03)报文中，当是CHAP认证方式时，必须出现在REQ_AUTH报文中。
        """
        for t,v in self.attrs:
            if t == 0x03:
                return pktutils.DecodeOctets(v)
                
    def get_chap_pwd(self):
        """
        长度: 16(固定)
        描述: 该属性表示由ppp chap用户通过MD5算法加密后的密码。
        当出现此属性时，其chap challenge在Challenge（03）属性中。当采用chap方式认证时，必须出现在REQ_AUTH（03）报文中。
        """
        for attr in self.attrs:
            if attr[0] == 0x04:
                return pktutils.DecodeOctets(attr[1])
                
    def get_text_info(self):
        """
        长度: 大于等于3,小于等于255
        描述: 该属性用于BAS将RADIUS等第三方认证设备的提示信息透传到Portal Server。
        当认证失败时，表示认证失败原因。
        必须出现在NTF_LOGOUT报文中，表示BAS强制用户下线的原因，当认证拒绝或者认证超时的时候，必须出现在ACK_AUTH报文中。
        长度至少为3字节，但不超过253字节。内容可以为任意字符串，但是不包括字符串结束符‘\0’。
        该属性可以存在于从BAS到Portal Server的任何报文中，同一个报文中允许有多个该属性，建议只携带1个。
        """
        texts = []
        for attr in self.attrs:
            if attr[0] == 0x05:
                texts.append(pktutils.DecodeString(attr[1]))
        return texts
                
    
    def get_up_link_flux(self):
        """
        长度: 2/10
        描述: 本属性只能用在REQ_INFO（Type＝9）和ACK_INFO（Type＝0x0a）报文中，数量不能超过1。
        当是REQ_INFO报文时，长度为2字节。
        当是ACK_INFO报文时，长度为10字节，表示该用户的上行（输出）的流量，
        用一个8字节（64Bits）无符号整数（网络顺序）表示，单位为kbytes。
        """
        pass
        
    def get_down_link_flux(self):
        """
        长度: 2/10
        本属性只能用在REQ_INFO（Type＝9）和ACK_INFO（Type＝0x0a）报文中，数量不能超过1。
        当是REQ_INFO报文时，长度为2字节。
        当是ACK_INFO报文时，长度为10字节，表示该用户的下行（输入）的流量，
        用一个8字节（64Bits）无符号整数（网络顺序）表示，单位为kbytes。
        """
        pass
        
    def get_port(self):
        """
        长度: >=2，<=255
        描述: 本属性只能用在REQ_INFO（Type＝9）和ACK_INFO（Type＝0x0a）报文中，数量不能超过1。
        当是REQ_INFO报文时，长度为2字节。
        当是ACK_INFO报文时，变长，但不能超过253字节。内容为一个字符串（无'\0'结束符）。
        属性值至少2字节，但不超过34字节。必须出现在REQ_INFO和ACK_INFO报文中。
        其格式为主机名-vlan-槽位（2 Bytes-vlan 标识（4 Bytes）@vlan-SSID-SSID标识(最长32字节)@SSID。
        """
        for attr in self.attrs:
            if attr[0] == 0x08:
                return pktutils.DecodeString(attr[1])
                
    def get_ip_config(self):
        """
        长度：4(固定)
        描述：用于二次地址方式中，表示用户IP的切换。
             当是二次地址方式时，必须出现在ACK_AUTH（0x03）、
             ACK_LOGOUT（0x06）、NTF_LOGOUT（0x08）和NTF_USERIPCHAN（0x0c）报文中。
             属性值4字节长，在不同报文类型中含义不同
             
        1,在ACK _AUTH（Type＝4）报文中，表示BAS设备端通知Portal Server此用户需要二次地址分配。
          属性值置为全1（0xFFFFFFFF），表示用户客户端必须后继触发DHCP相应的流程，
          Portal Server须将此消息通知用户客户端，触发DHCP过程释放私网IP，申请公网IP。
        2,在ACK_LOGOUT（Type＝6）和NTF_LOGOUT（Type＝8）报文中，
          表示用户此时使用的IP地址必须回收，Portal Server必须通知用户触发DHCP过程释放公网IP，
          设备将重新为用户分配一个私网的IP地址，属性值没有意义，置为全1。
        3,在NTF_USERIPCHANGE（Type＝0x0c）报文中，
          表示BAS通知Portal Server用户的IP地址更改。属性值为用户认证前的私网IP地址。
        
        """
        for attr in self.attrs:
            if attr[0] == 0x09:
                return pktutils.DecodeInteger(attr[1])
                
    def get_bas_ip(self):
        """
        长度: 4(固定)
        描述: 属性值4字节。只有在REQ_INFO、REQ_CHALLENGE报文中不含此属性，其他报文必须包含此属性。
        当处于IPv6环境时，其值为0，接入用户的IPv6地址将通过属性“BAS-IPv6”予以提供。
        """
        for attr in self.attrs:
            if attr[0] == 0x0a:
                return pktutils.DecodeAddress(attr[1])
                
    def get_session_id(self):
        """
        长度: 6(固定)
        描述: 此属性用来标识用户，建议取用户的mac地址。
        属性值6字节。当BAS关心用户的mac地址时，
        必须出现在ACK_AUTH，ACK_LOGOUT，NTF_LOGOUT和NTF_USERIPCHAN报文中，否则可以不出现。
        """
        for attr in self.attrs:
            if attr[0] == 0x0b:
                return pktutils.DecodeOctets(attr[1])
                
    def get_delay_time(self):
        """
        长度: 6(固定)
        描述: 用户记录报文的发送延时
        属性值6字节，用于REQ_LOGOUT（Type＝5）和NTF_LOGOUT（Type＝8）报文中，表示报文发送时间与实际发生时间的间隔。目前没有实现。
        """
        for attr in self.attrs:
            if attr[0] == 0x0c:
                return pktutils.DecodeOctets(attr[1])
                
    def get_user_list(self):
        """
        长度: >=6,<=254
        描述: 属性值最短4个字节，最长252个字节，
        用于用户心跳报文（NTF_USER_HEARTBEAT（Type=0x010））和用户心跳回应报文（ACK_NTF_USER_HEARTBEAT（Type=0x11））。
        属性值中包含了用户IP地址列表，每个用户IP地址占用4个字节，中间没有分隔符。
        """
        for attr in self.attrs:
            if attr[0] == 0x0d:
                return pktutils.DecodeOctets(attr[1])
                
    def get_eap_message(self):
        """
        长度: <=254
        描述: 此属性主要适用于EAP_TLS认证。允许出现多个，EAP认证方式时，必须出现在REQ_AUTH及NTF_CHALLENGE报文中。
        属性值最长255个字节，用于证书请求报文（NTF_CHALLENGE（Type=0x012））和认证请求报文（REQ_AUTH（Type=0x03））。
        REQ_AUTH报文中该属性值中需要包含登录名，NTF_CHALLENGE报文中该属性值主要用于传输证书。
        """
        texts = []
        for attr in self.attrs:
            if attr[0] == 0x0e:
                texts.append(pktutils.DecodeString(attr[1]))
        return texts
        
    def get_user_notify(self):
        """
        长度: <=254
        描述: 此属性主要用于透传的Radius计费回应报文中的hw_User_Notify内容。
        属性值最长255个字节，用于用户消息通知报文（NTF_USER_NOTIFY（Type=0x013））,
        将Radius服务器的计费开始回应报文中的hw_User_Notify内容透传给客户端。
        可以实现Portal认证通过后提示帐号余额等重要信息。
        """
        for attr in self.attrs:
            if attr[0] == 0x0f:
                return pktutils.DecodeOctets(attr[1])
                
    def get_bas_ipv6(self):
        """
        长度: 16
        描述: 用于标识BAS设备的IPv6地址，所有BAS设备发送的报文都应该携带该属性。
        属性值16字节。只有在REQ_INFO、REQ_CHALLENGE报文中不含此属性，其他报文必须包含此属性。IPv4环境时，其值为0。
        """
        for attr in self.attrs:
            if attr[0] == 0x64:
                return pktutils.DecodeOctets(attr[1])
                
    def get_useripv6_list(self):
        """
        长度: >=18,<=252
        描述: 用于标识接入用户的IPv6地址，其值由Portal Server根据用户获得的IPv6地址填写。
        属性值最短16个字节，最长252个字节，用于用户心跳报文（NTF_USER_HEARTBEAT（Type=0x010））
        和用户心跳回应报文（ACK_NTF_USER_HEARTBEAT（Type=0x11））。属性值中包含了用户IPv6地址列表，
        每个用户IPv6地址占用16个字节，中间没有分隔符。
        """
        for attr in self.attrs:
            if attr[0] == 0x65:
                return pktutils.DecodeOctets(attr[1])
        
    
    @property            
    def err_string(self):
        return ''
    
    @staticmethod           
    def newMessage(typ, userIp, serialNo,reqId,secret,
        auth='\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00'):
        return PortalV2(
            type = typ,
            userIp = pktutils.EncodeAddress(userIp),
            serialNo = serialNo,
            reqId = reqId,
            auth = auth,
            secret = six.b(secret)
        )
        
         
    @staticmethod        
    def newReqChallenge(userIp,secret,serialNo=None):
        """0x01"""
        pkt = PortalV2.newMessage(REQ_CHALLENGE,userIp,(serialNo or CurrentSN()),0,secret)
        pkt.auth_packet()
        return pkt
    
    @staticmethod         
    def newReqAuth(userIp,username,password,reqId,challenge,secret,basip,serialNo=None):
        """0x03"""
        pkt = PortalV2.newMessage(REQ_AUTH,userIp,(serialNo or CurrentSN()),reqId,secret)
        username = pktutils.EncodeString(username)
        password = pktutils.EncodeString(password)
        bas_addr = pktutils.EncodeAddress(basip)
        _reqid = struct.pack('>H',reqId)
        chap_pwd = md5_constructor("%s%s%s"%(_reqid[1],password,challenge)).digest()
        pkt.attrNum = 4
        pkt.attrs = [
            (0x01,username),
            (0x04,chap_pwd),
            (0x03,challenge),
            (0x0a,bas_addr)
        ]
        pkt.auth_packet()
        return pkt

        
    @staticmethod        
    def newReqLogout(userIp,secret,basip,serialNo=None):
        """0x05"""
        pkt = PortalV2.newMessage(
            REQ_LOGOUT,userIp,(serialNo or CurrentSN()),0,secret)
        bas_addr = pktutils.EncodeAddress(basip)
        pkt.attrNum = 1
        pkt.attrs = [
            (0x0a,bas_addr)
        ]
        pkt.auth_packet()
        return pkt
        
        
    @staticmethod        
    def newAffAckAuth(userIp,secret,basip,serialNo=None,reqId=None):
        """0x07"""
        pkt = PortalV2.newMessage(
            AFF_ACK_AUTH,userIp,(serialNo or CurrentSN()),(reqId or 0),secret)
        bas_addr = pktutils.EncodeAddress(basip)
        pkt.attrNum = 1
        pkt.attrs = [
            (0x0a,bas_addr)
        ]
        pkt.auth_packet()
        return pkt
        
        
    @staticmethod        
    def newReqInfo(userIp,secret,serialNo=None):
        """0x09"""
        pkt = PortalV2.newMessage(REQ_INFO,userIp,(serialNo or CurrentSN()),0,secret)
        pkt.attrNum = 1
        pkt.attrs = [(0x08,'\x00\x00')]
        pkt.auth_packet()
        return pkt  
        
    @staticmethod        
    def newNtfHeart(secret,basip):
        """0x0f NTF_HEARTBEAT"""
        pkt = PortalV2.newMessage(NTF_HEARTBEAT,'0.0.0.0',CurrentSN(),0,secret)
        bas_addr = pktutils.EncodeAddress(basip)
        pkt.attrNum = 1
        pkt.attrs = [
            (0x0a,bas_addr)
        ]
        pkt.auth_packet()
        return pkt

        
# XXX - ''.join([(len(`chr(x)`)==3) and chr(x) or '.' for x in range(256)])
__vis_filter = """................................ !"#$%&\'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[.]^_`abcdefghijklmnopqrstuvwxyz{|}~................................................................................................................................."""

def hexdump(buf, length=16):
    """Return a hexdump output string of the given buffer."""
    n = 0
    res = []
    while buf:
        line, buf = buf[:length], buf[length:]
        hexa = ' '.join(['%02x' % ord(x) for x in line])
        line = line.translate(__vis_filter)
        res.append('  %04d:  %-*s %s' % (n, length * 3, hexa, line))
        n += length
    return '\n'.join(res)
        
if __name__ == '__main__':
    pkt = PortalV2(packet="\x02\x02\x00\x00\x1b@\x00\x00\xac\x10\x01*\x00\x00\x01\x01\x06\xf2\xc3\xa0\xfe\xf1\x0c8\xea\xfb\x9aM\x86\xb4E\x06&\x06S\xf7O\xf6")
    print repr(pkt)

