#!/usr/bin/env python
#coding=utf-8
from toughradius.txradius.radius.packet import tools
from toughradius.txradius.radius.packet import AuthPacket
from toughradius.txradius.radius.packet import AcctPacket
from toughradius.txradius.radius.packet import CoAPacket
from toughradius.txradius.radius.packet import AccessRequest
from toughradius.txradius.radius.packet import AccessAccept
from toughradius.txradius.radius.packet import AccountingRequest
from toughradius.txradius.radius.packet import AccountingResponse
from toughradius.txradius.radius.packet import CoARequest
from toughradius.txradius.mschap import mschap,mppe
import time
import binascii
import datetime
import hashlib
from toughradius.common import six
import decimal

decimal.getcontext().prec = 16
decimal.getcontext().rounding = decimal.ROUND_UP


md5_constructor = hashlib.md5

PacketStatusTypeMap = {
    1 : 'AccessRequest',
    2 : 'AccessAccept',
    3 : 'AccessReject',
    4 : 'AccountingRequest',
    5 : 'AccountingResponse',
    40 : 'DisconnectRequest',
    41 : 'DisconnectACK',
    42 : 'DisconnectNAK',
    43 : 'CoARequest',
    44 : 'CoAACK',
    45 : 'CoANAK',
}



def format_packet_str(pkt):
    attr_keys = pkt.keys()
    _str = "\nRadius Packet:%s"%PacketStatusTypeMap.get(pkt.code)
    _str += "\nid:%s" % pkt.id
    _str += "\ncode:%s" % pkt.code
    _str += "\nauth:%s" % [pkt.authenticator]
    _str += "\nAttributes: "     
    for attr in attr_keys:
        try:
            _str += "\n\t%s: %s " % (attr, pkt[attr])               
            # _str += "\n\t%s: %s " % (attr, tools.DecodeAnyAttr(pkt[attr]))   
        except:
            pass
    return _str


def format_packet_log(pkt):
    attr_keys = pkt.keys()
    _str = "RadiusPacket:%s;" % PacketStatusTypeMap[pkt.code]
    _str += "id:%s; " % pkt.id
    _str += "code:%s; " % pkt.code
    _str += "auth:%s; " % [pkt.authenticator]
    for attr in attr_keys:
        try:
            _str += "%s:%s; " % (attr, pkt[attr])             
            # _str += "%s:%s; " % (attr, tools.DecodeAnyAttr(pkt[attr]))
        except:
            pass
    return _str


def get_session_timeout(pkt,defval=86400):
    try:return tools.DecodeInteger(pkt.get(27)[0]) or defval
    except:return defval

def get_interim_update(pkt,defval=300):
    try:return tools.DecodeInteger(pkt.get(85)[0]) or defval
    except:return defval

class ExtAttrMixin:

    def __init__(self):
        self._created = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        self._vendor_id = 0
        self._vlanid1 = 0
        self._vlanid2 = 0
        self._client_mac = None
        self._source = ('',0)
        self.resp_attrs = {}

    @property
    def source(self):
        return self._source

    @source.setter
    def source(self, source):
        self._source = source

    @property
    def vendor_id(self):
        return self._vendor_id

    @vendor_id.setter
    def vendor_id(self, vendor_id):
        self._vendor_id = vendor_id

    @property
    def vlanid1(self):
        return self._vlanid1

    @vlanid1.setter
    def vlanid1(self,vlanid1):
        self._vlanid1 = vlanid1

    @property
    def vlanid2(self):
        return self._vlanid2

    @vlanid2.setter
    def vlanid2(self,vlanid2):
        self._vlanid2 = vlanid2

    @property
    def client_mac(self):
        return self._client_mac

    @client_mac.setter
    def client_mac(self,macaddr):
        self._client_mac = macaddr

    def get_vlanids(self):
        return self.get_vlanid1(),self.get_vlanid2()  

    @property
    def created(self):
        return self._created



class CoAMessage(CoAPacket,ExtAttrMixin):
    def __init__(self, code=CoARequest, id=None, secret=six.b(''),
            authenticator=None, **attributes):
        CoAPacket.__init__(self, code, id, six.b(secret), authenticator, **attributes)
        ExtAttrMixin.__init__(self)
        
    def format_str(self):
        return format_packet_str(self)    

    def format_log(self):
        return format_packet_log(self)

    def get_acct_sessionid(self):
        try:return tools.DecodeString(self.get(44)[0])
        except:pass    

    def get_framed_ipaddr(self):
        try:return tools.DecodeAddress(self.get(8)[0])
        except:pass

    def get_nas_addr(self):
        try:
            return tools.DecodeAddress(self.get(4)[0])
        except:pass

class AuthMessage(AuthPacket,ExtAttrMixin):

    def __init__(self, code=AccessRequest, id=None, secret=six.b(''), authenticator=None, **attributes):
        AuthPacket.__init__(self, code, id, six.b(secret), authenticator, **attributes)
        ExtAttrMixin.__init__(self)

    def format_str(self):
        return format_packet_str(self)

    def format_log(self):
        return format_packet_log(self)

    def __str__(self):
        _str = PacketStatusTypeMap[self.code]
        _str += ",id=%s"%self.id
        if self.code == 1:
            _str += ",username=%s,mac_addr=%s" % (self.get_user_name(),self.get_mac_addr())
        if 'Reply-Message' in self:
            _str += ',Reply-Message="%s"' % self['Reply-Message'][0]
        return _str   

    def CreateReply(self, **attributes):
        return AuthMessage(AccessAccept, self.id,
            self.secret, self.authenticator, dict=self.dict,
            **attributes)
        
    def ChapEcrypt(self,password):
        if not self.authenticator:
            self.authenticator = self.CreateAuthenticator()
        if not self.id:
            self.id = self.CreateID()
        if isinstance(password, six.text_type):
            password = password.strip().encode('utf-8')

        chapid = self.authenticator[0]
        self['CHAP-Challenge'] = self.authenticator
        return '%s%s' % (chapid, md5_constructor("%s%s%s" % (chapid, password, self.authenticator)).digest())

   
    def get_nas_id(self):
        try:
            return tools.DecodeString(self.get(32)[0])
        except:pass

    def get_nas_portid(self):
        try:return tools.DecodeString(self.get(87)[0])
        except:return ''           

    def get_nas_port_type(self):
        try:return tools.DecodeInteger(self.get(61)[0]) or 0
        except:return 0

    def get_nas_addr(self):
        try:
            return tools.DecodeAddress(self.get(4)[0])
        except:pass

    def get_nas_class(self):
        try:return tools.DecodeString(self.get(25)[0])
        except:return ''
        
    def get_mac_addr(self):
        try:
            return self.client_mac or tools.DecodeString(self.get(31)[0]).replace("-",":")
        except:return None

    def get_user_name(self):
        try:
            user_name = tools.DecodeString(self.get(1)[0])
            return user_name
        except:
            return None

    def get_framed_ipaddr(self):
        try:return tools.DecodeAddress(self.get(8)[0])
        except:return ''

    def get_framed_netmask(self):
        try:return tools.DecodeAddress(self.get(9)[0])
        except:return ''

    def get_session_timeout(self):
        try:return tools.DecodeInteger(self.get(27)[0]) or 0
        except:return 0    

    def get_acct_interim_interval(self):
        try:return tools.DecodeInteger(self.get(85)[0]) or 0
        except:return 0    


    def get_domain(self):
        try:
            user_name = tools.DecodeString(self.get(1)[0])
            if "@" in user_name:
                return user_name[user_name.index("@")+1:]
        except:
            return None            
        

    def get_passwd(self):
        try:return self.PwDecrypt(self.get(2)[0])
        except:return None        

    def get_chappwd(self):
        try:return tools.DecodeOctets(self.get(3)[0])
        except:return None  
        
    def verifyChapEcrypt(self,userpwd):
        if isinstance(userpwd, six.text_type):
            userpwd = userpwd.strip().encode('utf-8')   

        _password = self.get_chappwd()
        if len(_password) != 17:
            return False

        chapid = _password[0]
        password = _password[1:]

        if not self.authenticator:
            self.authenticator = self.CreateAuthenticator()

        challenge = self.authenticator
        if 'CHAP-Challenge' in self:
            challenge = self['CHAP-Challenge'][0] 

        _pwd =  md5_constructor("%s%s%s"%(chapid,userpwd,challenge)).digest()
        return password == _pwd
        

    def verifyMsChapV2(self,userpwd):
        ms_chap_response = self['MS-CHAP2-Response'][0]
        authenticator_challenge = self['MS-CHAP-Challenge'][0]
        if len(ms_chap_response)!=50:
            raise Exception("Invalid MSCHAPV2-Response attribute length")
        # if isinstance(userpwd, six.text_type):
        #     userpwd = userpwd.strip().encode('utf-8')
        
        nt_response = ms_chap_response[26:50]
        peer_challenge = ms_chap_response[2:18]
        _user_name = self.get(1)[0]
        nt_resp = mschap.generate_nt_response_mschap2(
            authenticator_challenge,
            peer_challenge,
            _user_name,
            userpwd,
        )
        if nt_resp == nt_response:
            auth_resp = mschap.generate_authenticator_response(
                userpwd,
                nt_response,
                peer_challenge,
                authenticator_challenge,
                _user_name
            )
            self.resp_attrs['MS-CHAP2-Success'] = auth_resp
            self.resp_attrs['MS-MPPE-Encryption-Policy'] = '\x00\x00\x00\x01'
            self.resp_attrs['MS-MPPE-Encryption-Type'] = '\x00\x00\x00\x06'
            mppeSendKey, mppeRecvKey = mppe.mppe_chap2_gen_keys(userpwd, nt_response)
            send_key, recv_key = mppe.gen_radius_encrypt_keys(
                mppeSendKey,
                mppeRecvKey,
                self.secret,
                self.authenticator)
            self.resp_attrs['MS-MPPE-Send-Key'] = send_key
            self.resp_attrs['MS-MPPE-Recv-Key'] = recv_key
            return True
        else:
            self.resp_attrs['Reply-Message'] = "E=691 R=1 C=%s V=3 M=<password error>" % ('\0' * 32)
            return False
        
        
    def get_pwd_type(self):
        if 'MS-CHAP-Challenge' in self:
            if 'MS-CHAP-Response' in self:
                return 'mschapv1'
            elif 'MS-CHAP2-Response' in self:
                return 'mschapv2'
        elif 'CHAP-Password' in self:
            return 'chap'
        else:
            return 'pap'
            

    def is_valid_pwd(self,userpwd):
        pwd_type = self.get_pwd_type()
        print "radius password type %s" % pwd_type
        try:
            if pwd_type == 'pap':
                return userpwd == self.get_passwd()
            elif pwd_type == 'chap':
                return self.verifyChapEcrypt(userpwd)
            elif pwd_type == 'mschapv1':
                return False
            elif pwd_type == 'mschapv2':
                return self.verifyMsChapV2(userpwd)
            else:
                return False
        except Exception as err:
            import traceback
            traceback.print_exc()
            return False

    def PwDecrypt(self, password):
        buf = password
        pw = six.b('')

        last = self.authenticator
        while buf:
            hash = md5_constructor(self.secret + last).digest()
            for i in range(16):
                pw += chr(ord(hash[i]) ^ ord(buf[i]))

            (last, buf) = (buf[:16], buf[16:])

        while pw.endswith(six.b('\x00')):
            pw = pw[:-1]

        return pw.decode('utf-8')

    def PwCrypt(self, password):
        if self.authenticator is None:
            self.authenticator = self.CreateAuthenticator()

        if isinstance(password, six.text_type):
            password = password.encode('utf-8')

        buf = password
        if len(password) % 16 != 0:
            buf += six.b('\x00') * (16 - (len(password) % 16))

        hash = md5_constructor(self.secret + self.authenticator).digest()
        result = six.b('')

        last = self.authenticator
        while buf:
            hash = md5_constructor(self.secret + last).digest()
            for i in range(16):
                result += chr(ord(hash[i]) ^ ord(buf[i]))

            last = result[-16:]
            buf = buf[16:]

        return result

    @property
    def dict_message(self):
        return dict(
            nas_id = self.get_nas_id(),
            nas_addr = self.get_nas_addr(),
            username = self.get_user_name(),
            password = self.get_passwd(),
            mac_addr = self.get_mac_addr(),
            vlanid1 = self.vlanid1,
            vlanid2 = self.vlanid2,
            nas_port_id = self.get_nas_portid(),
            nas_port_type = self.get_nas_port_type(),
            nas_class = self.get_nas_class(),
            framed_ipaddr = self.get_framed_ipaddr(),
            framed_netmask = self.get_framed_netmask(),
            session_timeout = self.get_session_timeout(),
        )



class AcctMessage(AcctPacket,ExtAttrMixin):
    def __init__(self, code=AccountingRequest, id=None, secret=six.b(''),
            authenticator=None, **attributes):
        AcctPacket.__init__(self, code, id, six.b(secret), authenticator, **attributes)
        ExtAttrMixin.__init__(self)

    def format_str(self):
        return format_packet_str(self)

    def format_log(self):
        return format_packet_log(self)

    def __str__(self):
        _str = PacketStatusTypeMap.get(self.code)
        _str += ",id=%s"%self.id
        if self.code == 4:
            _str += ",username=%s,mac_addr=%s" % (self.get_user_name(),self.get_mac_addr())
        return _str   

    def CreateReply(self,**attributes):
        return AcctMessage(AccountingResponse, self.id,
            self.secret, self.authenticator, dict=self.dict,
            **attributes)    

    def get_user_name(self):
        try:
            return tools.DecodeString(self.get(1)[0])
            # if "@" in user_name:
            #     return user_name[:user_name.index("@")]
            # else:
            #     return user_name
        except:
            return None
 

    def get_mac_addr(self):
        try:
            return self.client_mac or tools.DecodeString(self.get(31)[0]).replace("-",":")
        except:return None

   
    def get_nas_id(self):
        try:
            return tools.DecodeString(self.get(32)[0])
        except:return ''

    def get_nas_addr(self):
        try:
            return tools.DecodeAddress(self.get(4)[0])
        except:return ''


    def get_nas_port(self):
        try:return tools.DecodeInteger(self.get(5)[0]) or 0
        except:return 0

    def get_service_type(self):
        try:return tools.DecodeInteger(self.get(0)[0]) or 0
        except:return 0
        
    def get_framed_ipaddr(self):
        try:return tools.DecodeAddress(self.get(8)[0])
        except:return ''

    def get_framed_netmask(self):
        try:return tools.DecodeAddress(self.get(9)[0])
        except:return ''

    def get_nas_class(self):
        try:return tools.DecodeString(self.get(25)[0])
        except:return ''   

    def get_session_timeout(self):
        try:return tools.DecodeInteger(self.get(27)[0]) or 0
        except:return 0

    def get_calling_stationid(self):
        try:return tools.DecodeString(self.get(31)[0])
        except:return ''   

    def get_acct_status_type(self):
        try:return tools.DecodeInteger(self.get(40)[0])
        except:return 0

    def get_acct_input_octets(self):
        try:return tools.DecodeInteger(self.get(42)[0]) or 0
        except:return 0

    def get_acct_output_octets(self):
        try:return tools.DecodeInteger(self.get(43)[0]) or 0
        except:return 0

    def get_acct_sessionid(self):
        try:return tools.DecodeString(self.get(44)[0])
        except:return ''                                                         

    def get_acct_sessiontime(self):
        try:return tools.DecodeInteger(self.get(46)[0]) or 0
        except:return 0                                                             

    def get_acct_input_packets(self):
        try:return tools.DecodeInteger(self.get(47)[0]) or 0
        except:return 0                                                       

    def get_acct_output_packets(self):
        try:return tools.DecodeInteger(self.get(48)[0]) or 0
        except:return 0           

    def get_acct_terminate_cause(self):
        try:return tools.DecodeInteger(self.get(49)[0]) or 0
        except:return 0           

    def get_acct_input_gigawords(self):
        try:return tools.DecodeInteger(self.get(52)[0]) or 0
        except:return 0       

    def get_acct_output_gigawords(self):
        try:return tools.DecodeInteger(self.get(53)[0]) or 0
        except:return 0    

    def get_input_total(self):
        bl = decimal.Decimal(self.get_acct_input_octets())/decimal.Decimal(1024)
        gl = decimal.Decimal(self.get_acct_input_gigawords())*decimal.Decimal(4*1024*1024)
        tl = bl + gl
        return int(tl.to_integral_value())   
        
    def get_output_total(self):
        bl = decimal.Decimal(self.get_acct_output_octets())/decimal.Decimal(1024)
        gl = decimal.Decimal(self.get_acct_output_gigawords())*decimal.Decimal(4*1024*1024)
        tl = bl + gl
        return int(tl.to_integral_value())                                                            

    def get_event_timestamp(self):
        try:
            return tools.DecodeDate(self.get(55)[0])
        except:
            return 0

    def get_event_timestamp_str(self,timetype=0):
        try:
            _time = tools.DecodeDate(self.get(55)[0])
            if timetype == 0:
                return datetime.datetime.fromtimestamp(_time).strftime("%Y-%m-%d %H:%M:%S")
            else:
                return datetime.datetime.fromtimestamp(_time-(8*3600)).strftime("%Y-%m-%d %H:%M:%S")
        except:
            return ''

    def get_nas_port_type(self):
        try:return tools.DecodeInteger(self.get(61)[0]) or 0
        except:return 0   

    def get_nas_portid(self):
        try:return tools.DecodeString(self.get(87)[0])
        except:return ''    

    def get_acct_start_time(self):
        dt = datetime.datetime.fromtimestamp(time.time()-self.get_acct_sessiontime())
        return dt.strftime("%Y-%m-%d %H:%M:%S")

    def get_ticket(self):
        return dict(
            username = self.get_user_name(),
            mac_addr = self.get_mac_addr(),
            nas_addr = self.get_nas_addr(),
            nas_port = self.get_nas_port(),
            service_type = self.get_service_type(),
            framed_ipaddr = self.get_framed_ipaddr(),
            framed_netmask = self.get_framed_netmask(),
            nas_class = self.get_nas_class(),
            session_timeout = self.get_session_timeout(),
            calling_station_id = self.get_calling_stationid(),
            acct_status_type = self.get_acct_status_type(),
            acct_input_octets = self.get_acct_input_octets(),
            acct_output_octets = self.get_acct_output_octets(),
            acct_session_id = self.get_acct_sessionid(),
            acct_session_time = self.get_acct_sessiontime(),
            acct_input_packets = self.get_acct_input_packets(),
            acct_output_packets = self.get_acct_output_packets(),
            acct_terminate_cause = self.get_acct_terminate_cause(),
            acct_input_gigawords = self.get_acct_input_gigawords(),
            acct_output_gigawords = self.get_acct_output_gigawords(),
            event_timestamp = self.get_event_timestamp_str(),
            nas_port_type=self.get_nas_port_type(),
            nas_port_id=self.get_nas_portid()
        )
        
    def get_billing(self):
        return dict(
            nas_id = self.get_nas_id(),
            acct_session_id = self.get_acct_sessionid(),
            username = self.get_user_name(),
            mac_addr = self.get_mac_addr(),
            nas_addr = self.get_nas_addr(),
            nas_port = self.get_nas_port(),
            nas_port_id = self.get_nas_portid(),
            nas_port_type = self.get_nas_port_type(),
            nas_class = self.get_nas_class(),
            framed_ipaddr = self.get_framed_ipaddr(),
            framed_netmask = self.get_framed_netmask(),
            session_timeout = self.get_session_timeout(),
            acct_input_total = self.get_input_total(),
            acct_output_total = self.get_output_total(),
            acct_start_time = self.get_acct_start_time(),
            acct_session_time = self.get_acct_sessiontime(),
            acct_input_packets = self.get_acct_input_packets(),
            acct_output_packets = self.get_acct_output_packets(),
            acct_terminate_cause = self.get_acct_terminate_cause(),
        )

    @property
    def dict_message(self):
        return self.get_billing()
