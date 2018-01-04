#coding=utf-8

from toughradius.modules import get_radius_attr
from toughradius.modules import parse_cisco_vlan
from toughradius.modules import VENDOR_ZTE
import logging

logger = logging.getLogger(__name__)

class RequestParse(object):

    parse_vlan = parse_cisco_vlan

    @staticmethod
    def parse_mac(req):
        mac_addr = get_radius_attr(req, 'Calling-Station-Id')
        if mac_addr:
            mac_addr = mac_addr[12:]
            _mac = (mac_addr[0:2], mac_addr[2:4], mac_addr[4:6], mac_addr[6:8], mac_addr[8:10], mac_addr[10:])
            req.client_mac = ':'.join(_mac)
        return req

    @staticmethod
    def handle_radius(req, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_ZTE:
                req = RequestParse.parse_mac(req)
                req = RequestParse.parse_vlan(req)
        except Exception as err:
            logger.error("request parse error {}".format(err.message),exc_info=debug)

        return req

class RateLimit(object):

    @staticmethod
    def handle_radius(req, reply, debug=False):
        try:
            if int(req.vendor_id) == VENDOR_ZTE:
                reply['ZTE-Rate-Ctrl-Scr-Up'] = int(reply.resp_attrs.get('input_rate', 0))
                reply['ZTE-Rate-Ctrl-Scr-Down'] = int(reply.resp_attrs.get('output_rate', 0))
        except Exception as err:
            logger.error("rate limit error {}".format(err.message),exc_info=debug)

        return reply


request_parse = RequestParse
rate_limit = RateLimit