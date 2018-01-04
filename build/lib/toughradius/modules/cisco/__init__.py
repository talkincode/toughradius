#coding=utf-8

from toughradius.modules import parse_cisco_vlan
from toughradius.modules import VENDOR_CISCO
import logging
logger = logging.getLogger(__name__)

class RequestParse(object):

    parse_vlan = parse_cisco_vlan

    @staticmethod
    def parse_mac(req):
        for attr in req:
            if attr not in 'Cisco-AVPair':
                continue
            attr_val = req[attr]
            if attr_val.startswith('client-mac-address'):
                mac_addr = attr_val[len("client-mac-address="):]
                mac_addr = mac_addr.replace('.', '')
                _mac = (mac_addr[0:2], mac_addr[2:4], mac_addr[4:6], mac_addr[6:8], mac_addr[8:10], mac_addr[10:])
                req.client_mac = ':'.join(_mac)
        return req


    @staticmethod
    def handle_radius(req, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_CISCO:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message),exc_info=debug)

        return req

request_parse = RequestParse